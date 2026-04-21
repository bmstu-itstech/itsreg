package bots_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/bmstu-itstech/itsreg/internal/domain/bots"
)

const ownerID = bots.UserID(1)

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

func buildSurveyScript() *bots.Script {
	start := bots.MustNewEntry("start", bots.MustNewState(1))
	return bots.MustNewScript(
		ownerID,
		"",
		[]bots.Node{greetingNode, fullNameNode, choosePillNode, redPillNode, bluePill},
		[]bots.Entry{start},
	)
}

func TestScript_EntryNProcess(t *testing.T) {
	script := buildSurveyScript()
	userID := bots.UserID(2)
	entryKey := bots.EntryKey("start")

	// Пользователь использует команду /start
	thread, msgs, err := script.Entry("bot-1", userID, entryKey)
	require.NoError(t, err)
	require.NotNil(t, thread)
	require.Equal(t, greetingNode.State(), thread.State())
	require.Equal(t, greetingNode.BotMessages(), msgs)

	// Пользователь вводит неожидаемый текст
	msgs, err = script.Process(thread, bots.MustNewMessage("мяу"))
	require.NoError(t, err)
	require.Empty(t, msgs)                                 // Нет новых сообщений
	require.Equal(t, greetingNode.State(), thread.State()) // Остаётся в том же состоянии
	require.Empty(t, thread.Answers())                     // Ответ не сохранился, так как не было перехода по ребру

	// Пользователь вводит "Далее" и переходит по ребру к узлу с ФИО
	msgs, err = script.Process(thread, bots.MustNewMessage("Далее"))
	require.NoError(t, err)
	require.Equal(t, fullNameNode.BotMessages(), msgs)
	require.Equal(t, fullNameNode.State(), thread.State())
	require.Empty(t, thread.Answers()) // Ответов пока нет, так как ребро имеет операцию noop

	// Пользователь возвращается назад
	msgs, err = script.Process(thread, bots.MustNewMessage("Назад"))
	require.NoError(t, err)
	require.Equal(t, greetingNode.BotMessages(), msgs)
	require.Equal(t, greetingNode.State(), thread.State())
	require.Empty(t, thread.Answers()) // Ответов нет, так как ребро "Назад" имеет операцию noop

	// Пользователь возвращается
	_, err = script.Process(thread, bots.MustNewMessage("Далее"))
	require.NoError(t, err)

	// Пользователь вводит своё имя
	msgs, err = script.Process(thread, bots.MustNewMessage("Иванов Иван"))
	require.NoError(t, err)
	require.Equal(t, choosePillNode.BotMessages(), msgs)
	require.Equal(t, choosePillNode.State(), thread.State())
	require.Equal(t, map[bots.State]bots.Message{
		bots.MustNewState(2): bots.MustNewMessage("Иванов Иван"),
	}, thread.Answers())

	// Пользователь выбирает красную таблетку
	msgs, err = script.Process(thread, bots.MustNewMessage("Красная"))
	require.NoError(t, err)
	require.Equal(t, redPillNode.BotMessages(), msgs)
	require.Equal(t, redPillNode.State(), thread.State())
	require.Equal(t, map[bots.State]bots.Message{
		bots.MustNewState(2): bots.MustNewMessage("Иванов Иван"),
		bots.MustNewState(3): bots.MustNewMessage("Красная"),
	}, thread.Answers())

	// Пользователь увидел реальность и передумал
	msgs, err = script.Process(thread, bots.MustNewMessage("Назад"))
	require.NoError(t, err)
	require.Equal(t, choosePillNode.BotMessages(), msgs)
	require.Equal(t, choosePillNode.State(), thread.State())
	require.Equal(t, map[bots.State]bots.Message{
		bots.MustNewState(2): bots.MustNewMessage("Иванов Иван"),
		bots.MustNewState(3): bots.MustNewMessage("Красная"),
	}, thread.Answers())

	// ... и выбрал синюю таблетку
	msgs, err = script.Process(thread, bots.MustNewMessage("Синяя"))
	require.NoError(t, err)
	require.Equal(t, bluePill.BotMessages(), msgs)
	require.Equal(t, bluePill.State(), thread.State())
	require.Equal(t, map[bots.State]bots.Message{
		bots.MustNewState(2): bots.MustNewMessage("Иванов Иван"),
		bots.MustNewState(3): bots.MustNewMessage("Красная\nСиняя"), // Запись второго ответа через Append
	}, thread.Answers())
}

func TestScript_EntryFailed(t *testing.T) {
	script := buildSurveyScript()
	userID := bots.UserID(2)
	key := bots.EntryKey("admin")

	_, _, err := script.Entry("bot-1", userID, key)
	require.ErrorIs(t, err, bots.ErrEntryNotFound)
	require.ErrorContains(t, err, string(key))
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
		_, err := bots.NewScript(ownerID, "", []bots.Node{node1, node2, node3}, []bots.Entry{entry})
		require.NoError(t, err)
	})

	t.Run("Non-existent node - invalid script", func(t *testing.T) {
		// Узел 1 имеет ребро к несуществующему узлу 3.
		entry := bots.MustNewEntry("start", bots.MustNewState(1))
		_, err := bots.NewScript(ownerID, "", []bots.Node{node1, node2}, []bots.Entry{entry})
		require.Error(t, err)
		requireValidationErrorDetails(t, err, []rawDetail{
			{"nodes", bots.ErrorCodeScriptNodeNotFound},
		})
		require.ErrorContains(t, err, "nodes[1]")
		require.ErrorContains(t, err, "nodes[3]")
	})

	t.Run("Non connected graph - invalid script", func(t *testing.T) {
		// Здесь хитрость. Вообще говоря граф из {1, 2, 3} является связным, и, казалось бы
		// ошибки здесь нет. Но у нас есть дополнительное условие - обход графа должен начинаться
		// с вершин, которые указаны в entries. Обходя граф с 3 узла мы остаёмся в 3 узле, а значит
		// скрипт не является связным.
		entry := bots.MustNewEntry("start", bots.MustNewState(3))
		_, err := bots.NewScript(ownerID, "", []bots.Node{node1, node2, node3}, []bots.Entry{entry})
		require.Error(t, err)
		requireValidationErrorDetails(t, err, []rawDetail{
			{"nodes", bots.ErrorCodeScriptNodeIsNotConnected},
			{"nodes", bots.ErrorCodeScriptNodeIsNotConnected},
		})
		require.ErrorContains(t, err, "nodes[1]")
		require.ErrorContains(t, err, "nodes[2]")
	})
}
