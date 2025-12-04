package http

import (
	"encoding/csv"
	"errors"
	"fmt"
	"net/http"

	"github.com/go-chi/render"

	"github.com/bmstu-itstech/itsreg-bots/internal/app"
	"github.com/bmstu-itstech/itsreg-bots/internal/app/dto"
	"github.com/bmstu-itstech/itsreg-bots/internal/app/dto/request"
	"github.com/bmstu-itstech/itsreg-bots/internal/app/port"
	"github.com/bmstu-itstech/itsreg-bots/internal/domain/bots"
	"github.com/bmstu-itstech/itsreg-bots/pkg/salad"
)

type Server struct {
	app *app.Application
}

func NewHTTPServer(app *app.Application) *Server {
	return &Server{app: app}
}

func (s *Server) GetBots(w http.ResponseWriter, r *http.Request) {
	bs, err := s.app.Queries.GetUserBots.Handle(r.Context(), request.GetUserBotsQuery{
		UserID: 1,
	})
	if err != nil {
		renderPlainError(w, r, err, http.StatusInternalServerError)
		return
	}

	render.JSON(w, r, batchBotsFromApp(bs))
}

func (s *Server) CreateBot(w http.ResponseWriter, r *http.Request) {
	req := PutBots{}
	if err := render.Decode(r, &req); err != nil {
		renderPlainError(w, r, err, http.StatusBadRequest)
		return
	}

	script, err := scriptToApp(req.Script)
	if err != nil {
		renderPlainError(w, r, err, http.StatusBadRequest)
		return
	}

	err = s.app.Commands.CreateBot.Handle(r.Context(), request.CreateBotCommand{
		BotID:  req.Id,
		Token:  req.Token,
		Author: 1,
		Script: script,
	})

	var iiErr bots.InvalidInputError
	if errors.As(err, &iiErr) {
		renderInvalidInputError(w, r, iiErr, http.StatusBadRequest)
		return
	}

	var mErr *bots.MultiError
	if errors.As(err, &mErr) {
		renderMultiError(w, r, mErr, http.StatusBadRequest)
		return
	}

	if err != nil {
		renderPlainError(w, r, err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Location", fmt.Sprintf("/bots/%s", req.Id))
	w.WriteHeader(http.StatusCreated)
}

func (s *Server) DeleteBot(w http.ResponseWriter, r *http.Request, id string) {
	err := s.app.Commands.DeleteBot.Handle(r.Context(), request.DeleteBotCommand{BotID: id})
	if errors.Is(err, port.ErrBotNotFound) {
		renderPlainError(w, r, err, http.StatusNotFound)
		return
	}
	if err != nil {
		renderPlainError(w, r, err, http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) GetAnswers(w http.ResponseWriter, r *http.Request, id string) {
	bot, err := s.app.Queries.GetBot.Handle(r.Context(), request.GetBotQuery{ID: id})
	if errors.Is(err, port.ErrBotNotFound) {
		renderPlainError(w, r, err, http.StatusNotFound)
		return
	}
	if err != nil {
		renderPlainError(w, r, err, http.StatusInternalServerError)
		return
	}

	threads, err := s.app.Queries.GetThreads.Handle(r.Context(), request.GetThreadsQuery{BotID: id})
	if err != nil {
		renderPlainError(w, r, err, http.StatusInternalServerError)
		return
	}

	err = renderCsvAnswers(w, bot.Script.Nodes, threads)
	if err != nil {
		renderPlainError(w, r, err, http.StatusInternalServerError)
		return
	}
}

func (s *Server) StartBot(w http.ResponseWriter, r *http.Request, id string) {
	err := s.app.Commands.Start.Handle(r.Context(), request.StartCommand{BotID: id})
	if errors.Is(err, port.ErrBotNotFound) {
		renderPlainError(w, r, err, http.StatusNotFound)
		return
	}
	if err != nil {
		renderPlainError(w, r, err, http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) StopBot(w http.ResponseWriter, r *http.Request, id string) {
	err := s.app.Commands.Stop.Handle(r.Context(), request.StopCommand{BotID: id})
	if errors.Is(err, port.ErrBotNotFound) {
		renderPlainError(w, r, err, http.StatusNotFound)
		return
	}
	if errors.Is(err, port.ErrRunningInstanceNotFound) {
		renderPlainError(w, r, err, http.StatusNotFound)
		return
	}
	if err != nil {
		renderPlainError(w, r, err, http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) DisableBot(w http.ResponseWriter, r *http.Request, botID string) {
	err := s.app.Commands.DisableBot.Handle(r.Context(), request.DisableBotCommand{BotID: botID})
	if errors.Is(err, port.ErrBotNotFound) {
		renderPlainError(w, r, err, http.StatusNotFound)
		return
	}
	if err != nil {
		renderPlainError(w, r, err, http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) EnableBot(w http.ResponseWriter, r *http.Request, botID string) {
	err := s.app.Commands.EnableBot.Handle(r.Context(), request.EnableBotCommand{BotID: botID})
	if errors.Is(err, port.ErrBotNotFound) {
		renderPlainError(w, r, err, http.StatusNotFound)
		return
	}
	if err != nil {
		renderPlainError(w, r, err, http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) GetBot(w http.ResponseWriter, r *http.Request, id string) {
	bot, err := s.app.Queries.GetBot.Handle(r.Context(), request.GetBotQuery{ID: id})
	if errors.Is(err, port.ErrBotNotFound) {
		renderPlainError(w, r, err, http.StatusNotFound)
		return
	}
	if err != nil {
		renderPlainError(w, r, err, http.StatusInternalServerError)
		return
	}

	render.JSON(w, r, botFromApp(bot))
}

func (s *Server) GetStatus(w http.ResponseWriter, r *http.Request, id string) {
	status, err := s.app.Queries.GetStatus.Handle(r.Context(), request.GetStatusQuery{BotID: id})
	if errors.Is(err, port.ErrBotNotFound) {
		renderPlainError(w, r, err, http.StatusNotFound)
		return
	}
	if err != nil {
		renderPlainError(w, r, err, http.StatusInternalServerError)
		return
	}

	render.JSON(w, r, Status(status))
}

func (s *Server) Mailing(w http.ResponseWriter, r *http.Request, botID string) {
	req := PostMailing{}
	if err := render.Decode(r, &req); err != nil {
		renderPlainError(w, r, err, http.StatusBadRequest)
		return
	}

	err := s.app.Commands.Mailing.Handle(r.Context(), request.MailingCommand{
		BotID:    botID,
		EntryKey: req.EntryKey,
		Users:    req.Users,
	})
	var mErr *bots.MultiError
	if errors.As(err, &mErr) {
		renderMultiError(w, r, mErr, http.StatusBadRequest)
		return
	}
	if errors.Is(err, port.ErrBotNotFound) {
		renderPlainError(w, r, err, http.StatusNotFound)
		return
	}
	if err != nil {
		renderPlainError(w, r, err, http.StatusInternalServerError)
		return
	}
}

const offset = 3

func renderCsvAnswers(w http.ResponseWriter, nodes []dto.Node, threads []dto.Thread) error {
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

func makeMapStateToIndex(threads []dto.Thread) map[int]int {
	states := make(map[int]bool)
	for _, thread := range threads {
		for state := range thread.Answers {
			states[state] = true
		}
	}

	sortedStates := make([]int, 0)
	for state := range states {
		sortedStates = salad.InsertSorted(sortedStates, state, func(x, y int) bool { return x < y })
	}

	m := make(map[int]int, len(sortedStates))
	for idx, state := range sortedStates {
		m[state] = idx
	}

	return m
}

const answerThreadIDHeadName = "#"
const answerTimestampHeadName = "Отметка времени"
const answerUsernameHeadName = "Никнейм"

func makeAnswersTHead(nodes []dto.Node, stateToIndex map[int]int) []string {
	head := make([]string, len(stateToIndex)+offset)

	head[0] = answerThreadIDHeadName
	head[1] = answerTimestampHeadName
	head[2] = answerUsernameHeadName

	for _, node := range nodes {
		idx, ok := stateToIndex[node.State]
		if ok {
			head[idx+offset] = node.Title
		}
	}

	return head
}

func makeAnswersTRow(thread dto.Thread, stateToIndex map[int]int) []string {
	row := make([]string, len(stateToIndex)+offset)

	row[0] = thread.ID
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

func makeAnswersTBody(threads []dto.Thread, stateToIndex map[int]int) [][]string {
	body := make([][]string, len(threads))
	for i, thread := range threads {
		body[i] = makeAnswersTRow(thread, stateToIndex)
	}
	return body
}
