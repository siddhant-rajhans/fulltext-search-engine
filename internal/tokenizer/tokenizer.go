package tokenizer

import (
	"strings"
	"unicode"
)

var stopWords = map[string]struct{}{
	"a": {}, "an": {}, "and": {}, "are": {}, "as": {}, "at": {},
	"be": {}, "but": {}, "by": {}, "for": {}, "if": {}, "in": {},
	"into": {}, "is": {}, "it": {}, "no": {}, "not": {}, "of": {},
	"on": {}, "or": {}, "such": {}, "that": {}, "the": {}, "their": {},
	"then": {}, "there": {}, "these": {}, "they": {}, "this": {},
	"to": {}, "was": {}, "will": {}, "with": {},
}

func Tokenize(text string) []string {
	text = strings.ToLower(text)

	var tokens []string
	var current strings.Builder

	for _, r := range text {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			current.WriteRune(r)
		} else if current.Len() > 0 {
			word := current.String()
			if _, stop := stopWords[word]; !stop && len(word) > 1 {
				tokens = append(tokens, word)
			}
			current.Reset()
		}
	}
	if current.Len() > 0 {
		word := current.String()
		if _, stop := stopWords[word]; !stop && len(word) > 1 {
			tokens = append(tokens, word)
		}
	}

	return tokens
}

func TokenizeWithPositions(text string) []TermPosition {
	text = strings.ToLower(text)

	var terms []TermPosition
	var current strings.Builder
	pos := 0

	for _, r := range text {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			current.WriteRune(r)
		} else if current.Len() > 0 {
			word := current.String()
			if _, stop := stopWords[word]; !stop && len(word) > 1 {
				terms = append(terms, TermPosition{Term: word, Position: pos})
				pos++
			}
			current.Reset()
		}
	}
	if current.Len() > 0 {
		word := current.String()
		if _, stop := stopWords[word]; !stop && len(word) > 1 {
			terms = append(terms, TermPosition{Term: word, Position: pos})
		}
	}

	return terms
}

type TermPosition struct {
	Term     string
	Position int
}
