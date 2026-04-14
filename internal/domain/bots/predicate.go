package bots

import (
	"fmt"
	"regexp"
)

const (
	ErrorCodeExactMatchPredicateEmptyText ErrorCode = "exact-match-predicate-empty-text"
	ErrorCodeRegexPredicateEmptyPattern   ErrorCode = "exact-match-predicate-empty-pattern"
	ErrorCodeRegexPredicateInvalidPattern ErrorCode = "regex-predicate-invalid-pattern"
)

type Predicate interface {
	Match(msg Message) bool
}

type AlwaysTruePredicate struct{}

func (p AlwaysTruePredicate) Match(_ Message) bool {
	return true
}

type ExactMatchPredicate struct {
	text string
}

func NewExactMatchPredicate(text string) (Predicate, error) {
	if text == "" {
		return nil, NewValidationError(NewValidationErrorDetail(
			"text", ErrorCodeExactMatchPredicateEmptyText, "text field in exact match predicate cannot be empty",
		))
	}
	return ExactMatchPredicate{text}, nil
}

func MustNewExactMatchPredicate(text string) Predicate {
	p, err := NewExactMatchPredicate(text)
	if err != nil {
		panic(err)
	}
	return p
}

func (p ExactMatchPredicate) Match(msg Message) bool {
	return p.text == msg.Text()
}

func (p ExactMatchPredicate) Text() string {
	return p.text
}

type RegexMatchPredicate struct {
	regex *regexp.Regexp
}

func NewRegexMatchPredicate(pattern string) (Predicate, error) {
	if pattern == "" {
		return nil, NewValidationError(NewValidationErrorDetail(
			"pattern", ErrorCodeRegexPredicateEmptyPattern, "pattern field in regex predicate cannot be empty",
		))
	}

	regex, err := regexp.Compile(pattern)
	if err != nil {
		return nil, NewValidationError(NewValidationErrorDetail(
			"pattern", ErrorCodeRegexPredicateInvalidPattern,
			fmt.Sprintf("failed to compile regex pattern %q: %s", pattern, err.Error()),
		))
	}

	return RegexMatchPredicate{regex}, nil
}

func MustNewRegexMatchPredicate(pattern string) Predicate {
	p, err := NewRegexMatchPredicate(pattern)
	if err != nil {
		panic(err)
	}
	return p
}

func (p RegexMatchPredicate) Match(msg Message) bool {
	return p.regex.MatchString(msg.Text())
}

func (p RegexMatchPredicate) Pattern() string {
	return p.regex.String()
}
