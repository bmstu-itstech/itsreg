package bots

import (
	"fmt"
	"regexp"
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
		return nil, NewInvalidInputError(
			"invalid-exact-match-predicate",
			"expected non-empty string for exact match predicate",
		)
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
		return nil, NewInvalidInputError(
			"invalid-regex-predicate",
			"expected not empty pattern for regexp predicate",
		)
	}

	regex, err := regexp.Compile(pattern)
	if err != nil {
		return nil, NewInvalidInputError(
			"invalid-regex-predicate-pattern",
			fmt.Sprintf("failed to compile regexp pattern: %s", pattern),
		)
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
