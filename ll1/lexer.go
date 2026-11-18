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

// Tokenize splits the input with the lexer of the assignment grammar
// (Default): see TokenizeFor.
func Tokenize(input string) ([]Token, error) {
	return TokenizeFor(Default(), input)
}

// TokenizeFor splits the input into tokens of grammar g. A word
// [A-Za-z_][A-Za-z0-9_]* equal to a terminal of g (a keyword such as "and"
// or "var") becomes that terminal and any other word becomes "id"; digit
// runs become "num"; a character that is a one-character terminal of g
// stands for itself; spaces are skipped and anything else is a
// *LexicalError. On success the last token is always "$".
func TokenizeFor(g *Grammar, input string) ([]Token, error) {
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
			word := string(runes[start:i])
			kind := TerminalID
			if g.isKeyword(word) {
				kind = word
			}
			tokens = append(tokens, Token{Kind: kind, Lexeme: word, Pos: start + 1})
		case isDigit(r):
			for i < len(runes) && isDigit(runes[i]) {
				i++
			}
			tokens = append(tokens, Token{Kind: TerminalNum, Lexeme: string(runes[start:i]), Pos: start + 1})
		case g.IsTerminal(string(r)):
			i++
			tokens = append(tokens, Token{Kind: string(r), Lexeme: string(r), Pos: start + 1})
		default:
			return nil, &LexicalError{Char: r, Pos: start + 1}
		}
	}
	tokens = append(tokens, Token{Kind: EndMarker, Lexeme: EndMarker, Pos: len(runes) + 1})
	return tokens, nil
}

// isKeyword reports whether word is a terminal of g on its own. The names
// "id" and "num" are not keywords: typed as words they are identifiers.
func (g *Grammar) isKeyword(word string) bool {
	return word != TerminalID && word != TerminalNum && g.IsTerminal(word)
}

// punctuation lists the one-character terminals of g that are not letters,
// in Terminals order: "+ - * / ( )" for the assignment grammar.
func (g *Grammar) punctuation() []string {
	var result []string
	for _, terminal := range g.Terminals {
		runes := []rune(terminal)
		if len(runes) == 1 && !isLetter(runes[0]) && !isDigit(runes[0]) {
			result = append(result, terminal)
		}
	}
	return result
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
