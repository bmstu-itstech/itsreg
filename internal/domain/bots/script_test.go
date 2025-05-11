package bots_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/bmstu-itstech/itsreg-bots/internal/domain/bots"
)

var (
	greetingNode = bots.MustNewNode(
		bots.State(1),
		[]bots.Edge{
			bots.NewEdge(bots.MustNewExactMatchPredicate("Далее"), bots.State(2), bots.NoOp{}),
		},
		[]bots.BotMessage{
			bots.MustNewBotMessage("Привет! Это бот-опросник", []bots.Option{"Далее"}),
		},
	)
	fullNameNode = bots.MustNewNode(
		bots.State(2),
		[]bots.Edge{
			bots.NewEdge(bots.MustNewExactMatchPredicate("Назад"), bots.State(1), bots.NoOp{}),
			bots.NewEdge(bots.AlwaysTruePredicate{}, bots.State(3), bots.SaveOp{}),
		},
		[]bots.BotMessage{
			bots.MustNewBotMessage("Продолжая пользоваться ботом, Вы подтверждаете...", []bots.Option{}),
			bots.MustNewBotMessage("Введите своё ФИО", []bots.Option{}),
		},
	)
	choosePillNode = bots.MustNewNode(
		bots.State(3),
		[]bots.Edge{
			bots.NewEdge(bots.MustNewExactMatchPredicate("Красная"), bots.State(10), bots.AppendOp{}),
			bots.NewEdge(bots.MustNewExactMatchPredicate("Синяя"), bots.State(11), bots.AppendOp{}),
			bots.NewEdge(bots.MustNewExactMatchPredicate("Назад"), bots.State(2), bots.NoOp{}),
		},
		[]bots.BotMessage{
			bots.MustNewBotMessage("Выбери таблетку:", []bots.Option{"Красная", "Синяя", "Назад"}),
		},
	)
	redPillNode = bots.MustNewNode(
		bots.State(10),
		[]bots.Edge{
			bots.NewEdge(bots.MustNewExactMatchPredicate("Назад"), bots.State(3), bots.NoOp{}),
		},
		[]bots.BotMessage{
			bots.MustNewBotMessage("Теперь ты увидел суровую реальность...", []bots.Option{}),
		},
	)
	bluePill = bots.MustNewNode(
		bots.State(11),
		[]bots.Edge{
			bots.NewEdge(bots.MustNewExactMatchPredicate("Назад"), bots.State(3), bots.NoOp{}),
		},
		[]bots.BotMessage{
			bots.MustNewBotMessage("Оставайся в иллюзии...", []bots.Option{}),
		},
	)
)

func buildSurveyScript() bots.Script {
	start := bots.MustNewEntry("start", bots.State(1))
	return bots.MustNewScript(
		[]bots.Node{greetingNode, fullNameNode, choosePillNode, redPillNode, bluePill},
		[]bots.Entry{start},
	)
}

func TestScript_EntryNProcess(t *testing.T) {
	script := buildSurveyScript()
	prtId := bots.NewParticipantId(1, "bot")
	prt := bots.MustNewParticipant(prtId, "username")

	// Пользователь нажимаем команду /start
	msgs, err := script.Entry(prt, "start")
	require.NoError(t, err)
	require.Equal(t, greetingNode.Messages(), msgs)
	thread, ok := prt.CurrentThread()
	require.True(t, ok)
	require.NotNil(t, thread)
	require.Equal(t, thread.State(), bots.State(1))

	// Пользователь вводит то, чего от него не ждут
	msgs, err = script.Process(prt, bots.NewMessage("/admin"))
	require.NoError(t, err)
	require.Empty(t, msgs)
	thread, ok = prt.CurrentThread()
	require.Equal(t, thread.State(), bots.State(1))
	require.Empty(t, thread.Answers())

	// Пользователь вводит то, что от него всё-таки ожидают
	msgs, err = script.Process(prt, bots.NewMessage("Далее"))
	require.NoError(t, err)
	require.Equal(t, fullNameNode.Messages(), msgs)
	thread, ok = prt.CurrentThread()
	require.Equal(t, thread.State(), bots.State(2))
	require.Empty(t, thread.Answers())

	// Пользователь вводит Назад
	msgs, err = script.Process(prt, bots.NewMessage("Назад"))
	require.NoError(t, err)
	require.Equal(t, greetingNode.Messages(), msgs)
	thread, ok = prt.CurrentThread()
	require.Equal(t, thread.State(), bots.State(1))
	require.Empty(t, thread.Answers())

	// Шагаем обратно
	msgs, err = script.Process(prt, bots.NewMessage("Далее"))
	require.NoError(t, err)
	require.Equal(t, fullNameNode.Messages(), msgs)
	thread, ok = prt.CurrentThread()
	require.Equal(t, thread.State(), bots.State(2))

	// Пользователь вводит своё ФИО
	msgs, err = script.Process(prt, bots.NewMessage("Иванов Иван Иванович"))
	require.NoError(t, err)
	require.Equal(t, choosePillNode.Messages(), msgs)
	thread, ok = prt.CurrentThread()
	require.Equal(t, thread.State(), bots.State(3))
	require.Equal(t, map[bots.State]bots.Message{
		bots.State(2): bots.NewMessage("Иванов Иван Иванович"),
	}, thread.Answers())

	// Пользователь выбирает красную таблетку
	msgs, err = script.Process(prt, bots.NewMessage("Красная"))
	require.NoError(t, err)
	require.Equal(t, redPillNode.Messages(), msgs)
	thread, ok = prt.CurrentThread()
	require.Equal(t, thread.State(), bots.State(10))
	require.Equal(t, map[bots.State]bots.Message{
		bots.State(2): bots.NewMessage("Иванов Иван Иванович"),
		bots.State(3): bots.NewMessage("Красная"),
	}, thread.Answers())

	// Пользователь увидел реальность и передумал
	msgs, err = script.Process(prt, bots.NewMessage("Назад"))
	require.NoError(t, err)
	require.Equal(t, choosePillNode.Messages(), msgs)
	thread, ok = prt.CurrentThread()
	require.Equal(t, thread.State(), bots.State(3))
	require.Equal(t, map[bots.State]bots.Message{
		bots.State(2): bots.NewMessage("Иванов Иван Иванович"),
		bots.State(3): bots.NewMessage("Красная"),
	}, thread.Answers())

	// ... и выбрал синюю таблетку
	msgs, err = script.Process(prt, bots.NewMessage("Синяя"))
	require.NoError(t, err)
	require.Equal(t, bluePill.Messages(), msgs)
	thread, ok = prt.CurrentThread()
	require.Equal(t, thread.State(), bots.State(11))
	require.Equal(t, map[bots.State]bots.Message{
		bots.State(2): bots.NewMessage("Иванов Иван Иванович"),
		bots.State(3): bots.NewMessage("Красная\nСиняя"),
	}, thread.Answers())
}

func TestScript_Entry(t *testing.T) {
	script := buildSurveyScript()
	prtId := bots.NewParticipantId(bots.UserId(1), "bot")
	prt := bots.MustNewParticipant(prtId, "username")

	_, err := script.Entry(prt, "admin")
	require.ErrorAs(t, err, &bots.EntryNotFoundError{})
	require.EqualError(t, err, "entry not found: admin")
}

func TestNewScript(t *testing.T) {
	node1 := bots.MustNewNode(bots.State(1), []bots.Edge{
		bots.NewEdge(bots.MustNewExactMatchPredicate("2"), bots.State(2), bots.NoOp{}),
		bots.NewEdge(bots.MustNewExactMatchPredicate("3"), bots.State(3), bots.NoOp{}),
	}, []bots.BotMessage{
		bots.NewBotMessageWithoutOptions("1"),
	})

	node2 := bots.MustNewNode(bots.State(2), []bots.Edge{
		bots.NewEdge(bots.MustNewExactMatchPredicate("2"), bots.State(2), bots.NoOp{}), // Цикл
		bots.NewEdge(bots.MustNewExactMatchPredicate("1"), bots.State(1), bots.NoOp{}), // Цикл на себя
	}, []bots.BotMessage{
		bots.NewBotMessageWithoutOptions("2"),
	})

	node3 := bots.MustNewNode(bots.State(3), []bots.Edge{}, []bots.BotMessage{
		bots.NewBotMessageWithoutOptions("3"),
	})

	t.Run("Valid script", func(t *testing.T) {
		entry := bots.MustNewEntry("start", bots.State(1))
		_, err := bots.NewScript([]bots.Node{node1, node2, node3}, []bots.Entry{entry})
		require.NoError(t, err)
	})

	t.Run("Non-existent node - invalid script", func(t *testing.T) {
		// Узел 1 имеет ребро к несуществующему узлу 3.
		entry := bots.MustNewEntry("start", bots.State(1))
		_, err := bots.NewScript([]bots.Node{node1, node2}, []bots.Entry{entry})
		require.Error(t, err)
		var ierr bots.InvalidInputError
		ok := errors.As(err, &ierr)
		require.True(t, ok)
		require.EqualError(t, err, "no node with state 3")
	})

	t.Run("Non-existent node - invalid script", func(t *testing.T) {
		// Здесь хитрость. Вообще говоря граф из {1, 2, 3} является связным, и, казалось бы
		// ошибки здесь нет. Но у нас есть дополнительное условие - обход графа должен начинаться
		// с вершин, которые указаны в entries. Обходя граф с 3 узла мы остаёмся в 3 узле, а значит
		// скрипт не является связным.
		entry := bots.MustNewEntry("start", bots.State(3))
		_, err := bots.NewScript([]bots.Node{node1, node2, node3}, []bots.Entry{entry})
		require.Error(t, err)
		var ierr bots.InvalidInputError
		ok := errors.As(err, &ierr)
		require.True(t, ok)
		require.ErrorContains(t, err, "not connected") // Порядок обхода мапы неизвестен
	})
}
