package diffcalc_test

import (
	"github.com/bmstu-itstech/itsreg-bots/pkg/diffcalc"
	"testing"

	"github.com/stretchr/testify/require"
)

type TestEntity struct {
	ID   int
	Name string
	Age  int
}

func TestCalculateChanges(t *testing.T) {
	t.Run("empty arrays", func(t *testing.T) {
		var before, after []TestEntity

		identity := func(a, b TestEntity) bool { return a.ID == b.ID }
		equal := func(a, b TestEntity) bool { return a == b }

		result := diffcalc.Changes(before, after, identity, equal)

		require.Empty(t, result.Added)
		require.Empty(t, result.Updated)
		require.Empty(t, result.Deleted)
	})

	t.Run("all added when before is empty", func(t *testing.T) {
		var before []TestEntity
		after := []TestEntity{
			{ID: 1, Name: "Alice"},
			{ID: 2, Name: "Bob"},
		}

		identity := func(a, b TestEntity) bool { return a.ID == b.ID }
		equal := func(a, b TestEntity) bool { return a == b }

		result := diffcalc.Changes(before, after, identity, equal)

		require.Len(t, result.Added, 2)
		require.Empty(t, result.Updated)
		require.Empty(t, result.Deleted)
		require.ElementsMatch(t, after, result.Added)
	})

	t.Run("all deleted when after is empty", func(t *testing.T) {
		before := []TestEntity{
			{ID: 1, Name: "Alice"},
			{ID: 2, Name: "Bob"},
		}
		var after []TestEntity

		identity := func(a, b TestEntity) bool { return a.ID == b.ID }
		equal := func(a, b TestEntity) bool { return a == b }

		result := diffcalc.Changes(before, after, identity, equal)

		require.Empty(t, result.Added)
		require.Empty(t, result.Updated)
		require.Len(t, result.Deleted, 2)
		require.ElementsMatch(t, before, result.Deleted)
	})

	t.Run("no changes when identical", func(t *testing.T) {
		before := []TestEntity{
			{ID: 1, Name: "Alice", Age: 25},
			{ID: 2, Name: "Bob", Age: 30},
		}
		after := []TestEntity{
			{ID: 1, Name: "Alice", Age: 25},
			{ID: 2, Name: "Bob", Age: 30},
		}

		identity := func(a, b TestEntity) bool { return a.ID == b.ID }
		equal := func(a, b TestEntity) bool { return a == b }

		result := diffcalc.Changes(before, after, identity, equal)

		require.Empty(t, result.Added)
		require.Empty(t, result.Updated)
		require.Empty(t, result.Deleted)
	})

	t.Run("mixed changes: added, updated, deleted", func(t *testing.T) {
		before := []TestEntity{
			{ID: 1, Name: "Alice", Age: 25},   // будет обновлен
			{ID: 2, Name: "Bob", Age: 30},     // будет удален
			{ID: 3, Name: "Charlie", Age: 35}, // останется без изменений
		}
		after := []TestEntity{
			{ID: 1, Name: "Alice Updated", Age: 26}, // обновленная версия
			{ID: 3, Name: "Charlie", Age: 35},       // без изменений
			{ID: 4, Name: "David", Age: 40},         // новый элемент
		}

		identity := func(a, b TestEntity) bool { return a.ID == b.ID }
		equal := func(a, b TestEntity) bool { return a == b }

		result := diffcalc.Changes(before, after, identity, equal)

		require.Len(t, result.Added, 1)
		require.Len(t, result.Updated, 1)
		require.Len(t, result.Deleted, 1)

		require.Equal(t, TestEntity{ID: 4, Name: "David", Age: 40}, result.Added[0])
		require.Equal(t, TestEntity{ID: 1, Name: "Alice Updated", Age: 26}, result.Updated[0])
		require.Equal(t, TestEntity{ID: 2, Name: "Bob", Age: 30}, result.Deleted[0])
	})

	t.Run("only updated when identity matches but content differs", func(t *testing.T) {
		before := []TestEntity{
			{ID: 1, Name: "Old Name", Age: 25},
		}
		after := []TestEntity{
			{ID: 1, Name: "New Name", Age: 26},
		}

		identity := func(a, b TestEntity) bool { return a.ID == b.ID }
		equal := func(a, b TestEntity) bool { return a == b }

		result := diffcalc.Changes(before, after, identity, equal)

		require.Empty(t, result.Added)
		require.Len(t, result.Updated, 1)
		require.Empty(t, result.Deleted)
		require.Equal(t, after[0], result.Updated[0])
	})

	t.Run("complex identity function", func(t *testing.T) {
		// Идентичность определяется по комбинации полей
		before := []TestEntity{
			{ID: 1, Name: "Alice", Age: 25},
			{ID: 2, Name: "Bob", Age: 30},
		}
		after := []TestEntity{
			{ID: 1, Name: "Alice", Age: 26}, // Age изменился, но ID+Name те же
			{ID: 3, Name: "Charlie", Age: 35},
		}

		// Идентичность по ID + Name
		identity := func(a, b TestEntity) bool {
			return a.ID == b.ID && a.Name == b.Name
		}
		equal := func(a, b TestEntity) bool { return a == b }

		result := diffcalc.Changes(before, after, identity, equal)

		require.Len(t, result.Added, 1)
		require.Len(t, result.Updated, 1)
		require.Len(t, result.Deleted, 1)

		require.Equal(t, TestEntity{ID: 3, Name: "Charlie", Age: 35}, result.Added[0])
		require.Equal(t, TestEntity{ID: 1, Name: "Alice", Age: 26}, result.Updated[0])
		require.Equal(t, TestEntity{ID: 2, Name: "Bob", Age: 30}, result.Deleted[0])
	})

	t.Run("custom equality function", func(t *testing.T) {
		before := []TestEntity{
			{ID: 1, Name: "Alice", Age: 25},
		}
		after := []TestEntity{
			{ID: 1, Name: "Alice", Age: 25}, // То же самое, но с разным Age
		}

		identity := func(a, b TestEntity) bool { return a.ID == b.ID }
		// Игнорируем Age при проверке равенства
		equal := func(a, b TestEntity) bool {
			return a.ID == b.ID && a.Name == b.Name
		}

		result := diffcalc.Changes(before, after, identity, equal)

		// Должно быть пусто, так как ID и Name одинаковые
		require.Empty(t, result.Added)
		require.Empty(t, result.Updated)
		require.Empty(t, result.Deleted)
	})
}
