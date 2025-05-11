package bots

// InvalidInputError возникает тогда и только тогда, когда ошибка
// возникла по причине невалидного ввода пользователя.
type InvalidInputError struct {
	slug string
	msg  string
}

func NewInvalidInputError(slug string, msg string) InvalidInputError {
	return InvalidInputError{
		slug: slug,
		msg:  msg,
	}
}

func (e InvalidInputError) Error() string {
	return e.msg
}

func (e InvalidInputError) Slug() string {
	return e.slug
}
