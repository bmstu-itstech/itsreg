package dto

type Predicate struct {
	Type string
	Data string // Любым образом сериализованные данные о предикате, зависит от Type
}
