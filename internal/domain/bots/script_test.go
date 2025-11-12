package bots_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/bmstu-itstech/itsreg-bots/internal/domain/bots"
)

var (
	greetingNode = bots.MustNewNode(
		bots.MustNewState(1),
		"Приветствие",
		[]bots.Edge{
			bots.NewEdge(bots.MustNewExactMatchPredicate("Далее"), bots.MustNewState(2), bots.NoOp{}),
		},
		[]bots.Message{
			bots.MustNewMessage("Привет! Это бот-опросник"),
		},
		[]bots.Option{
			bots.MustNewOption("Далее"),
		},
	)
	fullNameNode = bots.MustNewNode(
		bots.MustNewState(2),
		"ФИО",
		[]bots.Edge{
			bots.NewEdge(bots.MustNewExactMatchPredicate("Назад"), bots.MustNewState(1), bots.NoOp{}),
			bots.NewEdge(bots.AlwaysTruePredicate{}, bots.MustNewState(3), bots.SaveOp{}),
		},
		[]bots.Message{
			bots.MustNewMessage("Продолжая пользоваться ботом, Вы подтверждаете..."),
			bots.MustNewMessage("Введите своё ФИО"),
		},
		nil,
	)
	choosePillNode = bots.MustNewNode(
		bots.MustNewState(3),
		"Таблетка",
		[]bots.Edge{
			bots.NewEdge(bots.MustNewExactMatchPredicate("Красная"), bots.MustNewState(10), bots.AppendOp{}),
			bots.NewEdge(bots.MustNewExactMatchPredicate("Синяя"), bots.MustNewState(11), bots.AppendOp{}),
			bots.NewEdge(bots.MustNewExactMatchPredicate("Назад"), bots.MustNewState(2), bots.NoOp{}),
		},
		[]bots.Message{
			bots.MustNewMessage("Выбери таблетку:"),
		},
		[]bots.Option{
			bots.MustNewOption("Красная"),
			bots.MustNewOption("Синяя"),
			bots.MustNewOption("Назад"),
		},
	)
	redPillNode = bots.MustNewNode(
		bots.MustNewState(10),
		"Красная",
		[]bots.Edge{
			bots.NewEdge(bots.MustNewExactMatchPredicate("Назад"), bots.MustNewState(3), bots.NoOp{}),
		},
		[]bots.Message{
			bots.MustNewMessage("Теперь ты увидел суровую реальность..."),
		},
		nil,
	)
	bluePill = bots.MustNewNode(
		bots.MustNewState(11),
		"Синяя",
		[]bots.Edge{
			bots.NewEdge(bots.MustNewExactMatchPredicate("Назад"), bots.MustNewState(3), bots.NoOp{}),
		},
		[]bots.Message{
			bots.MustNewMessage("Оставайся в иллюзии..."),
		},
		nil,
	)
)

func buildSurveyScript() bots.Script {
	start := bots.MustNewEntry("start", bots.MustNewState(1))
	return bots.MustNewScript(
		[]bots.Node{greetingNode, fullNameNode, choosePillNode, redPillNode, bluePill},
		[]bots.Entry{start},
	)
}

func TestScript_EntryNProcess(t *testing.T) {
	script := buildSurveyScript()
	prtID := bots.NewParticipantID(1, "bot")
	prt := bots.MustNewParticipant(prtID)

	// Пользователь нажимаем команду /start
	msgs, err := script.Entry(prt, "start")
	require.NoError(t, err)
	require.Equal(t, greetingNode.BotMessages(), msgs)
	thread := prt.ActiveThread()
	require.NotNil(t, thread)
	require.Equal(t, thread.State(), bots.MustNewState(1))

	// Пользователь вводит то, чего от него не ждут
	msgs, err = script.Process(prt, bots.MustNewMessage("/admin"))
	require.NoError(t, err)
	require.Empty(t, msgs)
	thread = prt.ActiveThread()
	require.Equal(t, thread.State(), bots.MustNewState(1))
	require.Empty(t, thread.Answers())

	// Пользователь вводит то, что от него всё-таки ожидают
	msgs, err = script.Process(prt, bots.MustNewMessage("Далее"))
	require.NoError(t, err)
	require.Equal(t, fullNameNode.BotMessages(), msgs)
	thread = prt.ActiveThread()
	require.Equal(t, thread.State(), bots.MustNewState(2))
	require.Empty(t, thread.Answers())

	// Пользователь вводит Назад
	msgs, err = script.Process(prt, bots.MustNewMessage("Назад"))
	require.NoError(t, err)
	require.Equal(t, greetingNode.BotMessages(), msgs)
	thread = prt.ActiveThread()
	require.Equal(t, thread.State(), bots.MustNewState(1))
	require.Empty(t, thread.Answers())

	// Шагаем обратно
	msgs, err = script.Process(prt, bots.MustNewMessage("Далее"))
	require.NoError(t, err)
	require.Equal(t, fullNameNode.BotMessages(), msgs)
	thread = prt.ActiveThread()
	require.Equal(t, thread.State(), bots.MustNewState(2))

	// Пользователь вводит своё ФИО
	msgs, err = script.Process(prt, bots.MustNewMessage("Иванов Иван Иванович"))
	require.NoError(t, err)
	require.Equal(t, choosePillNode.BotMessages(), msgs)
	thread = prt.ActiveThread()
	require.Equal(t, thread.State(), bots.MustNewState(3))
	require.Equal(t, map[bots.State]bots.Message{
		bots.MustNewState(2): bots.MustNewMessage("Иванов Иван Иванович"),
	}, thread.Answers())

	// Пользователь выбирает красную таблетку
	msgs, err = script.Process(prt, bots.MustNewMessage("Красная"))
	require.NoError(t, err)
	require.Equal(t, redPillNode.BotMessages(), msgs)
	thread = prt.ActiveThread()
	require.Equal(t, thread.State(), bots.MustNewState(10))
	require.Equal(t, map[bots.State]bots.Message{
		bots.MustNewState(2): bots.MustNewMessage("Иванов Иван Иванович"),
		bots.MustNewState(3): bots.MustNewMessage("Красная"),
	}, thread.Answers())

	// Пользователь увидел реальность и передумал
	msgs, err = script.Process(prt, bots.MustNewMessage("Назад"))
	require.NoError(t, err)
	require.Equal(t, choosePillNode.BotMessages(), msgs)
	thread = prt.ActiveThread()
	require.Equal(t, thread.State(), bots.MustNewState(3))
	require.Equal(t, map[bots.State]bots.Message{
		bots.MustNewState(2): bots.MustNewMessage("Иванов Иван Иванович"),
		bots.MustNewState(3): bots.MustNewMessage("Красная"),
	}, thread.Answers())

	// ... и выбрал синюю таблетку
	msgs, err = script.Process(prt, bots.MustNewMessage("Синяя"))
	require.NoError(t, err)
	require.Equal(t, bluePill.BotMessages(), msgs)
	thread = prt.ActiveThread()
	require.Equal(t, thread.State(), bots.MustNewState(11))
	require.Equal(t, map[bots.State]bots.Message{
		bots.MustNewState(2): bots.MustNewMessage("Иванов Иван Иванович"),
		bots.MustNewState(3): bots.MustNewMessage("Красная\nСиняя"),
	}, thread.Answers())
}

func TestScript_Entry(t *testing.T) {
	script := buildSurveyScript()
	prtID := bots.NewParticipantID(bots.UserID(1), "bot")
	prt := bots.MustNewParticipant(prtID)

	_, err := script.Entry(prt, "admin")
	require.ErrorAs(t, err, &bots.EntryNotFoundError{})
	require.EqualError(t, err, "entry not found: admin")
}

func TestNewScript(t *testing.T) {
	node1 := bots.MustNewNode(bots.MustNewState(1), "node1", []bots.Edge{
		bots.NewEdge(bots.MustNewExactMatchPredicate("2"), bots.MustNewState(2), bots.NoOp{}),
		bots.NewEdge(bots.MustNewExactMatchPredicate("3"), bots.MustNewState(3), bots.NoOp{}),
	}, []bots.Message{
		bots.MustNewMessage("1"),
	}, nil)

	node2 := bots.MustNewNode(bots.MustNewState(2), "node2", []bots.Edge{
		bots.NewEdge(bots.MustNewExactMatchPredicate("2"), bots.MustNewState(2), bots.NoOp{}), // Цикл
		bots.NewEdge(bots.MustNewExactMatchPredicate("1"), bots.MustNewState(1), bots.NoOp{}), // Цикл на себя
	}, []bots.Message{
		bots.MustNewMessage("2"),
	}, nil)

	node3 := bots.MustNewNode(bots.MustNewState(3), "node3", []bots.Edge{}, []bots.Message{
		bots.MustNewMessage("3"),
	}, nil)

	t.Run("Valid script", func(t *testing.T) {
		entry := bots.MustNewEntry("start", bots.MustNewState(1))
		_, err := bots.NewScript([]bots.Node{node1, node2, node3}, []bots.Entry{entry})
		require.NoError(t, err)
	})

	t.Run("Non-existent node - invalid script", func(t *testing.T) {
		// Узел 1 имеет ребро к несуществующему узлу 3.
		entry := bots.MustNewEntry("start", bots.MustNewState(1))
		_, err := bots.NewScript([]bots.Node{node1, node2}, []bots.Entry{entry})
		require.Error(t, err)
		var iiErr bots.InvalidInputError
		require.ErrorAs(t, err, &iiErr)
		require.Equal(t, "node-not-found", iiErr.Code)
		require.Contains(t, iiErr.Details, "state")
		require.Equal(t, "1", iiErr.Details["state"])
	})

	t.Run("Non-existent node - invalid script", func(t *testing.T) {
		// Здесь хитрость. Вообще говоря граф из {1, 2, 3} является связным, и, казалось бы
		// ошибки здесь нет. Но у нас есть дополнительное условие - обход графа должен начинаться
		// с вершин, которые указаны в entries. Обходя граф с 3 узла мы остаёмся в 3 узле, а значит
		// скрипт не является связным.
		entry := bots.MustNewEntry("start", bots.MustNewState(3))
		_, err := bots.NewScript([]bots.Node{node1, node2, node3}, []bots.Entry{entry})
		require.Error(t, err)
		var ierr bots.InvalidInputError
		ok := errors.As(err, &ierr)
		require.True(t, ok)
		var iiErr bots.InvalidInputError
		require.ErrorAs(t, err, &iiErr)
		require.Equal(t, "node-is-not-connected", iiErr.Code)
		require.Contains(t, iiErr.Details, "state")
		// Какой именно state - неизвестно, порядок обхода map не определён.
	})
}
