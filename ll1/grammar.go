// Package ll1 implementa o analisador LL(1) preditivo tabular (não
// recursivo): conjuntos FIRST e FOLLOW, tabela de parsing, analisador léxico,
// analisador com pilha e gerador de sentenças. A gramática do trabalho é a
// das expressões aritméticas; as outras gramáticas prontas ficam em presets.go.
package ll1

import "strings"

// Epsilon é o símbolo da cadeia vazia.
const Epsilon = "ε"

// EndMarker marca o fim da entrada e o fundo da pilha.
const EndMarker = "$"

// Nomes dos terminais que o léxico dá a identificadores e números.
const (
	TerminalID  = "id"
	TerminalNum = "num"
)

// Arrow separa a cabeça do corpo quando uma produção é impressa.
const Arrow = " -> "

// Production é uma regra numerada da gramática.
type Production struct {
	Number int      // começa em 1 e é a posição em Grammar.Productions
	Head   string   // não terminal do lado esquerdo
	Body   []string // []string{Epsilon} na produção vazia
}

// IsEmpty diz se a produção deriva a cadeia vazia diretamente (A -> ε).
func (p Production) IsEmpty() bool {
	return len(p.Body) == 1 && p.Body[0] == Epsilon
}

// String imprime a produção como "E' -> + T E'".
func (p Production) String() string {
	return p.Head + Arrow + strings.Join(p.Body, " ")
}

// Grammar é uma gramática livre de contexto com os símbolos na ordem de exibição.
type Grammar struct {
	Start        string
	NonTerminals []string // ordem de exibição: E E' T T' F
	Terminals    []string // id num + - * / ( )   (sem ε e sem $)
	Productions  []Production
}

// Default devolve a gramática fixa das expressões, a do trabalho.
func Default() *Grammar {
	bodies := []struct {
		head string
		body string
	}{
		{"E", "T E'"},
		{"E'", "+ T E'"},
		{"E'", "- T E'"},
		{"E'", Epsilon},
		{"T", "F T'"},
		{"T'", "* F T'"},
		{"T'", "/ F T'"},
		{"T'", Epsilon},
		{"F", "( E )"},
		{"F", TerminalID},
		{"F", TerminalNum},
	}
	g := &Grammar{
		Start:        "E",
		NonTerminals: []string{"E", "E'", "T", "T'", "F"},
		Terminals:    []string{TerminalID, TerminalNum, "+", "-", "*", "/", "(", ")"},
	}
	for i, b := range bodies {
		g.Productions = append(g.Productions, Production{
			Number: i + 1,
			Head:   b.head,
			Body:   strings.Fields(b.body),
		})
	}
	return g
}

// IsTerminal diz se s é um terminal da gramática.
func (g *Grammar) IsTerminal(s string) bool {
	return contains(g.Terminals, s)
}

// IsNonTerminal diz se s é um não terminal da gramática.
func (g *Grammar) IsNonTerminal(s string) bool {
	return contains(g.NonTerminals, s)
}

// ProductionsOf devolve, em ordem, as produções cuja cabeça é nonTerminal.
func (g *Grammar) ProductionsOf(nonTerminal string) []Production {
	var result []Production
	for _, p := range g.Productions {
		if p.Head == nonTerminal {
			result = append(result, p)
		}
	}
	return result
}

// String imprime uma linha por não terminal, com as alternativas juntas:
// "E' -> + T E' | - T E' | ε".
func (g *Grammar) String() string {
	var sb strings.Builder
	for _, nt := range g.NonTerminals {
		var alternatives []string
		for _, p := range g.ProductionsOf(nt) {
			alternatives = append(alternatives, strings.Join(p.Body, " "))
		}
		sb.WriteString(nt)
		sb.WriteString(Arrow)
		sb.WriteString(strings.Join(alternatives, " | "))
		sb.WriteString("\n")
	}
	return sb.String()
}

func contains(list []string, s string) bool {
	for _, item := range list {
		if item == s {
			return true
		}
	}
	return false
}
