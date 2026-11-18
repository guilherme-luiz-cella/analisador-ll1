package ll1

import "fmt"

// Table é a tabela de parsing LL(1) M[não terminal, terminal].
type Table struct {
	Grammar   *Grammar
	Columns   []string                  // Terminals + "$"
	Cells     map[string]map[string]int // não terminal -> coluna -> número da produção (ausente = erro)
	Conflicts []string                  // vazio numa gramática LL(1)

	// ConflictCells lista cada célula pedida por mais de uma produção, na
	// ordem em que os conflitos apareceram.
	ConflictCells []ConflictCell
}

// ConflictCell é uma célula da tabela pedida por mais de uma produção.
type ConflictCell struct {
	NonTerminal string
	Column      string
	Numbers     []int // todas as produções que pedem a célula; a primeira fica com ela
}

// BuildTable preenche a tabela a partir de FIRST e FOLLOW. Uma célula pedida
// duas vezes é um conflito; a primeira produção fica com a célula.
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
			// Regra 1 da tabela: para A → α, cada a ∈ FIRST(α) põe A → α em M[A, a].
			// Regra 2 da tabela: se ε ∈ FIRST(α), cada b ∈ FOLLOW(A), inclusive $,
			// põe A → α em M[A, b].
			if bodyFirst[column] || (bodyFirst[Epsilon] && follow[p.Head][column]) {
				t.place(p, column)
			}
		}
	}
	return t
}

// place escreve a produção na célula. Célula já ocupada quer dizer que a
// gramática não é LL(1): a primeira produção fica e o conflito é anotado.
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

// recordConflict junta numa só entrada todas as produções que disputam a
// mesma célula, para a tela mostrar o choque inteiro.
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

// Production devolve a produção em M[nonTerminal, column], se houver.
func (t *Table) Production(nonTerminal, column string) (Production, bool) {
	number, ok := t.Cells[nonTerminal][column]
	if !ok {
		return Production{}, false
	}
	// Os números começam em 1 e seguem a ordem de Productions.
	return t.Grammar.Productions[number-1], true
}
