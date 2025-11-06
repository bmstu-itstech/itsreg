package http

import (
	"encoding/csv"
	"errors"
	"fmt"
	"net/http"

	"github.com/go-chi/render"

	"github.com/bmstu-itstech/itsreg-bots/internal/app"
	"github.com/bmstu-itstech/itsreg-bots/internal/domain/bots"
	"github.com/bmstu-itstech/itsreg-bots/pkg/salad"
)

type Server struct {
	app *app.Application
}

func (s *Server) GetBots(w http.ResponseWriter, r *http.Request) {
	bs, err := s.app.Queries.GetUserBots.Handle(r.Context(), app.GetUserBots{
		UserId: 1,
	})
	if err != nil {
		httpError(w, r, err, http.StatusInternalServerError)
		return
	}

	render.JSON(w, r, batchBotsFromApp(bs))
}

func (s *Server) CreateBot(w http.ResponseWriter, r *http.Request) {
	req := PutBots{}
	if err := render.Decode(r, &req); err != nil {
		httpError(w, r, err, http.StatusBadRequest)
		return
	}

	script, err := scriptToApp(req.Script)
	if err != nil {
		httpError(w, r, err, http.StatusBadRequest)
		return
	}

	err = s.app.Commands.CreateBot.Handle(r.Context(), app.CreateBot{
		BotId:  req.Id,
		Token:  req.Token,
		Author: 1,
		Script: script,
	})

	var iiErr bots.InvalidInputError
	if errors.As(err, &iiErr) {
		httpSlugError(w, r, iiErr.Error(), iiErr.Slug(), http.StatusBadRequest)
		return
	}
	if err != nil {
		httpError(w, r, err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Location", fmt.Sprintf("/bots/%s", req.Id))
	w.WriteHeader(http.StatusCreated)
}

func (s *Server) GetAnswers(w http.ResponseWriter, r *http.Request, id string) {
	bot, err := s.app.Queries.GetBot.Handle(r.Context(), app.GetBot{Id: id})
	if errors.Is(err, bots.ErrBotNotFound) {
		httpError(w, r, err, http.StatusNotFound)
		return
	}
	if err != nil {
		httpError(w, r, err, http.StatusInternalServerError)
		return
	}

	threads, err := s.app.Queries.GetThreads.Handle(r.Context(), app.GetThreads{BotId: id})
	if err != nil {
		httpError(w, r, err, http.StatusInternalServerError)
		return
	}

	err = renderCsvAnswers(w, bot.Script.Nodes, threads)
	if err != nil {
		httpError(w, r, err, http.StatusInternalServerError)
		return
	}
}

func (s *Server) StartBot(w http.ResponseWriter, r *http.Request, id string) {
	err := s.app.Commands.Start.Handle(r.Context(), app.Start{BotId: id})
	if errors.Is(err, bots.ErrBotNotFound) {
		httpError(w, r, err, http.StatusNotFound)
		return
	}
	if err != nil {
		httpError(w, r, err, http.StatusInternalServerError)
		return
	}
}

func (s *Server) StopBot(w http.ResponseWriter, r *http.Request, id string) {
	err := s.app.Commands.Stop.Handle(r.Context(), app.Stop{BotId: id})
	if errors.Is(err, bots.ErrBotNotFound) {
		httpError(w, r, err, http.StatusNotFound)
		return
	}
	if errors.Is(err, bots.ErrRunningInstanceNotFound) {
		httpError(w, r, err, http.StatusNotFound)
		return
	}
	if err != nil {
		httpError(w, r, err, http.StatusInternalServerError)
		return
	}
}

func (s *Server) GetBot(w http.ResponseWriter, r *http.Request, id string) {
	bot, err := s.app.Queries.GetBot.Handle(r.Context(), app.GetBot{Id: id})
	if errors.Is(err, bots.ErrBotNotFound) {
		httpError(w, r, err, http.StatusNotFound)
		return
	}
	if err != nil {
		httpError(w, r, err, http.StatusInternalServerError)
		return
	}

	render.JSON(w, r, botFromApp(bot))
}

func NewHTTPServer(app *app.Application) *Server {
	return &Server{app: app}
}

func httpError(w http.ResponseWriter, r *http.Request, err error, code int) {
	render.Status(r, code)
	render.JSON(w, r, Error{Message: err.Error()})
}

func httpSlugError(w http.ResponseWriter, r *http.Request, msg string, slug string, code int) {
	render.Status(r, code)
	render.JSON(w, r, Error{
		Message: msg,
		Slug:    &slug,
	})
}

const offset = 3

func renderCsvAnswers(w http.ResponseWriter, nodes []app.Node, threads []app.Thread) error {
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")

	utf8bom := []byte{0xEF, 0xBB, 0xBF}
	_, _ = w.Write(utf8bom)

	writer := csv.NewWriter(w)

	stateToIndex := makeMapStateToIndex(threads)
	thead := makeAnswersTHead(nodes, stateToIndex)
	tbody := makeAnswersTBody(threads, stateToIndex)

	if err := writer.Write(thead); err != nil {
		return fmt.Errorf("failed to write CSV answers table: %w", err)
	}
	if err := writer.WriteAll(tbody); err != nil {
		return fmt.Errorf("failed to write CSV answers table: %w", err)
	}

	return nil
}

func makeMapStateToIndex(threads []app.Thread) map[uint]int {
	states := make(map[uint]bool)
	for _, thread := range threads {
		for state := range thread.Answers {
			states[state] = true
		}
	}

	sortedStates := make([]uint, 0)
	for state := range states {
		sortedStates = salad.InsertSorted(sortedStates, state, func(x, y uint) bool { return x < y })
	}

	m := make(map[uint]int, len(sortedStates))
	for idx, state := range sortedStates {
		m[state] = idx
	}

	return m
}

const answerThreadIDHeadName = "#"
const answerTimestampHeadName = "Отметка времени"
const answerUsernameHeadName = "Никнейм"

func makeAnswersTHead(nodes []app.Node, stateToIndex map[uint]int) []string {
	head := make([]string, len(stateToIndex)+offset)

	head[0] = answerThreadIDHeadName
	head[1] = answerTimestampHeadName
	head[2] = answerUsernameHeadName

	for _, node := range nodes {
		idx, ok := stateToIndex[node.State]
		fmt.Printf("%v\n", node)
		if ok {
			head[idx+offset] = node.Title
		}
	}

	return head
}

func makeAnswersTRow(thread app.Thread, stateToIndex map[uint]int) []string {
	row := make([]string, len(stateToIndex)+offset)

	row[0] = thread.Id
	row[1] = thread.StartedAt.Format("2006-01-02 15:04:05")
	row[2] = thread.Username

	for state, ans := range thread.Answers {
		idx, ok := stateToIndex[state]
		if ok {
			row[idx+offset] = ans.Text
		}
	}

	return row
}

func makeAnswersTBody(threads []app.Thread, stateToIndex map[uint]int) [][]string {
	body := make([][]string, len(threads))
	for i, thread := range threads {
		body[i] = makeAnswersTRow(thread, stateToIndex)
	}
	return body
}
