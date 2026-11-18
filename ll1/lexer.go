package ll1

import "fmt"

// Token é uma unidade léxica da entrada.
type Token struct {
	Kind   string // nome do terminal ("id", "num", "+", ...) ou "$"
	Lexeme string // texto como foi digitado ("soma", "42", "+"); "$" no fim
	Pos    int    // coluna na entrada, contando a partir de 1
}

// LexicalError aponta um caractere que não começa nenhum token.
type LexicalError struct {
	Char rune
	Pos  int // coluna na entrada, contando a partir de 1
}

// Error devolve a mensagem completa, com o prefixo "erro léxico".
func (e *LexicalError) Error() string {
	return "erro léxico: " + e.Detail()
}

// Detail descreve o caractere inválido e onde ele está.
func (e *LexicalError) Detail() string {
	return fmt.Sprintf("caractere '%c' inválido na posição %d", e.Char, e.Pos)
}

// Tokenize separa a entrada com o léxico da gramática do trabalho (Default);
// veja TokenizeFor.
func Tokenize(input string) ([]Token, error) {
	return TokenizeFor(Default(), input)
}

// TokenizeFor separa a entrada em tokens da gramática g. Uma palavra
// [A-Za-z_][A-Za-z0-9_]* igual a um terminal de g (palavra-chave como "and"
// ou "var") vira esse terminal, e qualquer outra palavra vira "id"; uma
// sequência de dígitos vira "num"; um caractere que é terminal de g vale por
// si mesmo; espaços são pulados e qualquer outra coisa é *LexicalError. Sem
// erro, o último token é sempre "$".
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

// isKeyword diz se word é, sozinha, um terminal de g. Os nomes "id" e "num"
// não são palavras-chave: digitados como palavras, são identificadores.
func (g *Grammar) isKeyword(word string) bool {
	return word != TerminalID && word != TerminalNum && g.IsTerminal(word)
}

// punctuation lista os terminais de um caractere de g que não são letra nem
// dígito, na ordem de Terminals: "+ - * / ( )" na gramática do trabalho.
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
