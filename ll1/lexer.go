package ll1

import "fmt"

// Token is one lexical unit of the input.
type Token struct {
	Kind   string // terminal name ("id", "num", "+", ...) or "$"
	Lexeme string // text as typed ("soma", "42", "+"); "$" for the end
	Pos    int    // 1-based column in the input
}

// LexicalError reports a character that does not start any token.
type LexicalError struct {
	Char rune
	Pos  int // 1-based column in the input
}

func (e *LexicalError) Error() string {
	return "erro léxico: " + e.Detail()
}

// Detail describes the bad character and where it is.
func (e *LexicalError) Detail() string {
	return fmt.Sprintf("caractere '%c' inválido na posição %d", e.Char, e.Pos)
}

// operators are the single-character terminals; each is its own kind.
const operators = "+-*/()"

// Tokenize splits the input into tokens. Identifiers [A-Za-z_][A-Za-z0-9_]*
// become "id", digit runs become "num", spaces are skipped and any other
// character is a *LexicalError. On success the last token is always "$".
func Tokenize(input string) ([]Token, error) {
	runes := []rune(input)
	var tokens []Token
	for i := 0; i < len(runes); {
		r := runes[i]
		start := i
		switch {
		case isSpace(r):
			i++
			continue
		case isLetter(r):
			for i < len(runes) && (isLetter(runes[i]) || isDigit(runes[i])) {
				i++
			}
			tokens = append(tokens, Token{Kind: TerminalID, Lexeme: string(runes[start:i]), Pos: start + 1})
		case isDigit(r):
			for i < len(runes) && isDigit(runes[i]) {
				i++
			}
			tokens = append(tokens, Token{Kind: TerminalNum, Lexeme: string(runes[start:i]), Pos: start + 1})
		case isOperator(r):
			i++
			tokens = append(tokens, Token{Kind: string(r), Lexeme: string(r), Pos: start + 1})
		default:
			return nil, &LexicalError{Char: r, Pos: start + 1}
		}
	}
	tokens = append(tokens, Token{Kind: EndMarker, Lexeme: EndMarker, Pos: len(runes) + 1})
	return tokens, nil
}

func isSpace(r rune) bool {
	return r == ' ' || r == '\t' || r == '\r' || r == '\n'
}

func isLetter(r rune) bool {
	return r == '_' || (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z')
}

func isDigit(r rune) bool {
	return r >= '0' && r <= '9'
}

func isOperator(r rune) bool {
	for _, op := range operators {
		if r == op {
			return true
		}
	}
	return false
}
