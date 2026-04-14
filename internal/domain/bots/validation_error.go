package bots

import (
	"errors"
	"strings"
)

const fieldDelimiter = "."

type ErrorCode string

const ErrorCodeUnknown ErrorCode = "unknown"

// ValidationErrorDetail является частью ошибки ValidationError
//
// Не реализует интерфейс Error, так как не является самостоятельной ошибкой.
type ValidationErrorDetail struct {
	Field   string
	Code    ErrorCode
	Message string
}

func NewValidationErrorDetail(field string, code ErrorCode, message string) ValidationErrorDetail {
	return ValidationErrorDetail{Field: field, Code: code, Message: message}
}

func (e ValidationErrorDetail) WithPrefix(prefix string) ValidationErrorDetail {
	return ValidationErrorDetail{
		Field:   prefix + fieldDelimiter + e.Field,
		Code:    e.Code,
		Message: e.Message,
	}
}

type ValidationError struct {
	Details []ValidationErrorDetail
}

// NewValidationError создаёт ошибку из деталей.
func NewValidationError(details ...ValidationErrorDetail) ValidationError {
	return ValidationError{
		Details: details,
	}
}

func (e ValidationError) Error() string {
	var builder strings.Builder
	builder.WriteString("validation error:")
	for _, detail := range e.Details {
		builder.WriteString(" ")
		builder.WriteString(detail.Message)
		builder.WriteString(";")
	}
	return builder.String()
}

func (e ValidationError) IsZero() bool {
	return len(e.Details) == 0
}

// WithPrefix создаёт копию ошибки, где для каждой детали ошибки к Field добавляет prefix, используя
// fieldDelimiter в качестве разделителя.
//
//	d := ValidationErrorDetail{ "text", ErrorCodeEmptyTitle, "message title can't be empty" }
//	err := NewValidationError(d)
//	err := err.WithPrefix("message[2]")
//	err.Details[0].Field	// "message[2].text"
func (e ValidationError) WithPrefix(prefix string) ValidationError {
	details := make([]ValidationErrorDetail, len(e.Details))
	for i, detail := range e.Details {
		details[i] = detail.WithPrefix(prefix)
	}
	return ValidationError{Details: details}
}

// Merge объединяет детали двух ошибок в одну. Не изменяет исходные ошибки, а
// возвращает новую.
func (e ValidationError) Merge(o ValidationError) ValidationError {
	return ValidationError{
		Details: append(e.Details, o.Details...),
	}
}

// AppendPrefixed добавляет детали из ошибки err с префиксом к текущему набору деталей.
// Если err не является ValidationError, создаёт деталь с кодом ErrorCodeUnknown.
// Используется для рекурсивного наращивания префикса при валидации вложенных структур.
//
//	var vErr bots.ValidationError
//	if err := ...; err != nil {
//		vErr = vErr.AppendPrefixed(err, "field")
//	}
func (e ValidationError) AppendPrefixed(err error, prefix string) ValidationError {
	if err == nil {
		return e
	}

	var vErr ValidationError
	if errors.As(err, &vErr) {
		return e.Merge(vErr.WithPrefix(prefix))
	}

	detail := NewValidationErrorDetail(prefix, ErrorCodeUnknown, err.Error())
	return e.Merge(NewValidationError(detail))
}

// AppendError добавляет детали из err к текущему набору деталей e.
// Если err не является ValidationError, создаёт деталь с кодом ErrorCodeUnknown.
func (e ValidationError) AppendError(err error) ValidationError {
	if err == nil {
		return e
	}

	var vErr ValidationError
	if errors.As(err, &vErr) {
		return e.Merge(vErr)
	}

	detail := NewValidationErrorDetail("", ErrorCodeUnknown, err.Error())
	return e.Merge(NewValidationError(detail))
}

// OrError возвращает ошибку, если детали есть, иначе nil.
// Используется вместо проверки `if len(details) > 0`.
//
//	var details []ValidationErrorDetail
//	// ... накопление деталей ...
//	if err := NewValidationError(details...).OrError(); err != nil {
//		return nil, err
//	}
func (e ValidationError) OrError() error {
	if len(e.Details) == 0 {
		return nil
	}
	return e
}
