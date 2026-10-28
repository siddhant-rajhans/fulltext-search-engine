package tokenizer

import (
	"testing"
)

func TestTokenizeBasic(t *testing.T) {
	tokens := Tokenize("The quick brown fox jumps over the lazy dog")
	expected := []string{"quick", "brown", "fox", "jumps", "over", "lazy", "dog"}
	assertTokens(t, expected, tokens)
}

func TestTokenizeRemovesStopWords(t *testing.T) {
	tokens := Tokenize("this is a test of the system")
	expected := []string{"test", "system"}
	assertTokens(t, expected, tokens)
}

func TestTokenizeLowercases(t *testing.T) {
	tokens := Tokenize("Hello WORLD GoLang")
	for _, tok := range tokens {
		for _, r := range tok {
			if r >= 'A' && r <= 'Z' {
				t.Fatalf("expected lowercase, got %s", tok)
			}
		}
	}
}

func TestTokenizeHandlesPunctuation(t *testing.T) {
	tokens := Tokenize("hello, world! testing... one-two-three")
	for _, tok := range tokens {
		for _, r := range tok {
			if !((r >= 'a' && r <= 'z') || (r >= '0' && r <= '9')) {
				t.Fatalf("unexpected character in token: %s", tok)
			}
		}
	}
}

func TestTokenizeEmptyString(t *testing.T) {
	tokens := Tokenize("")
	if len(tokens) != 0 {
		t.Fatalf("expected empty, got %v", tokens)
	}
}

func TestTokenizeWithPositions(t *testing.T) {
	terms := TokenizeWithPositions("quick brown fox quick")
	if len(terms) != 4 {
		t.Fatalf("expected 4 terms, got %d", len(terms))
	}
	if terms[0].Position != 0 || terms[0].Term != "quick" {
		t.Fatalf("expected (quick,0), got (%s,%d)", terms[0].Term, terms[0].Position)
	}
	if terms[3].Position != 3 || terms[3].Term != "quick" {
		t.Fatalf("expected (quick,3), got (%s,%d)", terms[3].Term, terms[3].Position)
	}
}

func TestSingleCharTokensDropped(t *testing.T) {
	tokens := Tokenize("I a x go in")
	// "I", "a", "x" are single chars or stop words; "go" is 2 chars; "in" is stop word
	for _, tok := range tokens {
		if len(tok) < 2 {
			t.Fatalf("single-char token should be dropped: %s", tok)
		}
	}
}

func assertTokens(t *testing.T, expected, actual []string) {
	t.Helper()
	if len(expected) != len(actual) {
		t.Fatalf("expected %d tokens, got %d: %v", len(expected), len(actual), actual)
	}
	for i := range expected {
		if expected[i] != actual[i] {
			t.Fatalf("token[%d]: expected %q, got %q", i, expected[i], actual[i])
		}
	}
}
