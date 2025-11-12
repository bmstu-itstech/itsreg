package bots

import (
	"errors"
	"fmt"
	"strings"
)

type InvalidInputError struct {
	Code    string
	Message string
	Details map[string]string
}

func NewInvalidInputError(code string, message string, context ...string) InvalidInputError {
	details := make(map[string]string)
	for i := 0; i < len(context); i += 2 {
		if i+1 < len(context) {
			details[context[i]] = context[i+1]
		}
	}
	return InvalidInputError{Code: code, Message: message, Details: details}
}

func (e InvalidInputError) Error() string {
	var b strings.Builder
	for k, v := range e.Details {
		b.WriteString(fmt.Sprintf(" %s='%s'", k, v))
	}
	return fmt.Sprintf("%s: %s %s", e.Code, e.Message, strings.TrimSpace(b.String()))
}

type MultiError struct {
	Errors []error
}

func (e *MultiError) Error() string {
	return fmt.Sprintf("multiple errors: %v", e.Errors)
}

func (e *MultiError) Append(err error) {
	e.Errors = append(e.Errors, err)
}

func (e *MultiError) Extend(o *MultiError) {
	e.Errors = append(e.Errors, o.Errors...)
}

// ExtendOrAppend расширяет список ошибок, если err является MultiError; иначе добавляет единичную ошибку.
func (e *MultiError) ExtendOrAppend(err error) {
	var mErr *MultiError
	if errors.As(err, &mErr) {
		e.Extend(mErr)
	} else {
		e.Append(err)
	}
}

func (e *MultiError) HasError() bool {
	return len(e.Errors) > 0
}
