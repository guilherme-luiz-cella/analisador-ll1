// Package ll1 implements a table-driven (non-recursive) predictive LL(1)
// parser for a fixed arithmetic expression grammar: FIRST and FOLLOW sets,
// the parsing table, a lexer, the stack parser and a sentence generator.
package ll1

import "strings"

// Epsilon is the empty string symbol.
const Epsilon = "ε"

// EndMarker marks the end of the input and the bottom of the stack.
const EndMarker = "$"

// Terminal names produced by the lexer for identifiers and numbers.
const (
	TerminalID  = "id"
	TerminalNum = "num"
)

// Arrow separates the head from the body when a production is printed.
const Arrow = " -> "

// Production is one numbered grammar rule.
type Production struct {
	Number int      // 1-based
	Head   string   // non-terminal on the left side
	Body   []string // []string{Epsilon} for the empty production
}

// IsEmpty reports whether the production derives the empty string directly.
func (p Production) IsEmpty() bool {
	return len(p.Body) == 1 && p.Body[0] == Epsilon
}

// String prints the production as "E' -> + T E'".
func (p Production) String() string {
	return p.Head + Arrow + strings.Join(p.Body, " ")
}

// Grammar is a context-free grammar with symbols kept in display order.
type Grammar struct {
	Start        string
	NonTerminals []string // display order: E E' T T' F
	Terminals    []string // id num + - * / ( )   (no ε, no $)
	Productions  []Production
}

// Default returns the fixed expression grammar of the assignment.
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

// IsTerminal reports whether s is one of the grammar's terminals.
func (g *Grammar) IsTerminal(s string) bool {
	return contains(g.Terminals, s)
}

// IsNonTerminal reports whether s is one of the grammar's non-terminals.
func (g *Grammar) IsNonTerminal(s string) bool {
	return contains(g.NonTerminals, s)
}

// ProductionsOf returns the productions whose head is nonTerminal, in order.
func (g *Grammar) ProductionsOf(nonTerminal string) []Production {
	var result []Production
	for _, p := range g.Productions {
		if p.Head == nonTerminal {
			result = append(result, p)
		}
	}
	return result
}

// String prints one line per non-terminal with alternatives grouped:
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
