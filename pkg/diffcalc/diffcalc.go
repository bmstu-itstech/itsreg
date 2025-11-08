package diffcalc

type ChangeSet[T any] struct {
	Added   []T
	Updated []T
	Deleted []T
}

func (s ChangeSet[T]) IsZero() bool {
	return len(s.Added) == 0 && len(s.Updated) == 0 && len(s.Deleted) == 0
}

// Changes возвращает набор изменений (добавление, изменение, удаление), который должен быть проведён над слайсом
// before, чтобы провести его к виду after.
// Функция identity определяет идентичность объекта, то есть свойство сохранения своей уникальности
// после изменений. Функция equal определяет полное равенство объектов.
// Так, сущности, имеют идентичность по их ID, а объекты значения - по набору всех полей; равенство в обоих случаях
// определяется набором всех полей.
func Changes[T any](before, after []T, identity, equal func(lhs, rhs T) bool) ChangeSet[T] {
	var r ChangeSet[T]

	for _, b := range before {
		if a, ok := find(after, b, identity); !ok {
			r.Deleted = append(r.Deleted, b)
		} else if !equal(b, a) {
			r.Updated = append(r.Updated, a)
		}
	}

	for _, a := range after {
		if _, ok := find(before, a, identity); !ok {
			r.Added = append(r.Added, a)
		}
	}

	return r
}

func find[T any](slice []T, item T, identity func(lsh, rhs T) bool) (T, bool) {
	for _, v := range slice {
		if identity(item, v) {
			return v, true
		}
	}
	var zero T
	return zero, false
}

func Equal[T comparable](lhs, rhs T) bool {
	return lhs == rhs
}
