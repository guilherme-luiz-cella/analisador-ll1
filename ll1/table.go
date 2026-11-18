package ll1

import "fmt"

// Table is the LL(1) parsing table M[non-terminal, terminal].
type Table struct {
	Grammar   *Grammar
	Columns   []string                  // Terminals + "$"
	Cells     map[string]map[string]int // non-terminal -> column -> production Number (missing = error)
	Conflicts []string                  // empty for an LL(1) grammar

	// ConflictCells lists each cell claimed by more than one production,
	// in the order the conflicts were found.
	ConflictCells []ConflictCell
}

// ConflictCell is a table cell that more than one production claims.
type ConflictCell struct {
	NonTerminal string
	Column      string
	Numbers     []int // every production claiming the cell; the first keeps it
}

// BuildTable fills the table from FIRST and FOLLOW: production A -> α goes
// in M[A, a] for every terminal a in FIRST(α), and, when ε is in FIRST(α),
// in M[A, b] for every b in FOLLOW(A). A cell claimed twice is a conflict;
// the first production keeps the cell.
func BuildTable(g *Grammar) *Table {
	t := &Table{
		Grammar: g,
		Columns: append(append([]string{}, g.Terminals...), EndMarker),
		Cells:   make(map[string]map[string]int, len(g.NonTerminals)),
	}
	for _, nt := range g.NonTerminals {
		t.Cells[nt] = map[string]int{}
	}
	first := g.firstSets()
	follow := g.followSets(first)
	for _, p := range g.Productions {
		bodyFirst := g.firstOfSequence(first, p.Body)
		for _, column := range t.Columns {
			if bodyFirst[column] || (bodyFirst[Epsilon] && follow[p.Head][column]) {
				t.place(p, column)
			}
		}
	}
	return t
}

func (t *Table) place(p Production, column string) {
	row := t.Cells[p.Head]
	if existing, taken := row[column]; taken {
		t.Conflicts = append(t.Conflicts, fmt.Sprintf(
			"Conflito em M[%s, %s]: produções %d e %d", p.Head, column, existing, p.Number))
		t.recordConflict(p, column, existing)
		return
	}
	row[column] = p.Number
}

func (t *Table) recordConflict(p Production, column string, existing int) {
	for i, cell := range t.ConflictCells {
		if cell.NonTerminal == p.Head && cell.Column == column {
			t.ConflictCells[i].Numbers = append(cell.Numbers, p.Number)
			return
		}
	}
	t.ConflictCells = append(t.ConflictCells, ConflictCell{
		NonTerminal: p.Head,
		Column:      column,
		Numbers:     []int{existing, p.Number},
	})
}

// Production returns the production in M[nonTerminal, column], if any.
func (t *Table) Production(nonTerminal, column string) (Production, bool) {
	number, ok := t.Cells[nonTerminal][column]
	if !ok {
		return Production{}, false
	}
	for _, p := range t.Grammar.Productions {
		if p.Number == number {
			return p, true
		}
	}
	return Production{}, false
}
