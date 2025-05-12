package http

import (
	"encoding/csv"
	"errors"
	"fmt"
	"net/http"
	"slices"

	"github.com/go-chi/render"

	"github.com/bmstu-itstech/itsreg-bots/internal/app"
	"github.com/bmstu-itstech/itsreg-bots/internal/domain/bots"
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

	err = renderCsvAnswers(w, bot.Script, threads)
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

func renderCsvAnswers(w http.ResponseWriter, script app.Script, threads []app.Thread) error {
	writer := csv.NewWriter(w)
	w.Header().Set("Content-Type", "text/csv")

	thead, stateToIndex := answersTHead(script)
	tbody := answersTBody(threads, stateToIndex)

	if err := writer.Write(thead); err != nil {
		return err
	}
	if err := writer.WriteAll(tbody); err != nil {
		return err
	}

	return nil
}

const answerThreadIDHeadName = "#"
const answerTimestampHeadName = "Отметка времени"
const answerUsernameHeadName = "Никнейм"

func answersTHead(script app.Script) ([]string, map[uint]int) {
	nodes := script.Nodes
	slices.SortFunc(nodes, func(a, b app.Node) int { return int(a.State) - int(b.State) })

	head := make([]string, 0, len(nodes)+3) // ThreadID + Username + Timestamp
	stateToIndex := make(map[uint]int)

	head = append(head, answerThreadIDHeadName, answerTimestampHeadName, answerUsernameHeadName)
	for i, node := range nodes {
		head = append(head, node.Title)
		stateToIndex[node.State] = i + 3 // Смещение на три столбца
	}

	return head, stateToIndex
}

func answersTRow(thread app.Thread, stateToIndex map[uint]int) []string {
	row := make([]string, len(stateToIndex)) // Индексы идут подряд
	row[0] = thread.Id
	row[1] = thread.StartedAt.Format("2006-01-02 15:04:05")
	row[2] = thread.Username
	for state, ans := range thread.Answers {
		i, ok := stateToIndex[state]
		if ok {
			row[i] = ans.Text
		}
	}
	return row
}

func answersTBody(threads []app.Thread, stateToIndex map[uint]int) [][]string {
	body := make([][]string, len(threads))
	for i, thread := range threads {
		body[i] = answersTRow(thread, stateToIndex)
	}
	return body
}
