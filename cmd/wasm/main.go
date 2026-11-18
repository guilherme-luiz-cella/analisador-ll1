//go:build js && wasm

// Command wasm exposes the ll1 package to the browser as globalThis.ll1.
// Every function returns a JSON string, except generate and generateInvalid,
// which return the sentence itself.
package main

import (
	"encoding/json"
	"math/rand"
	"strings"
	"syscall/js"
	"time"

	"analisador-ll1/ll1"
)

// maxGenerateDepth keeps generated sentences short enough to read on a projector.
const maxGenerateDepth = 4

type productionJSON struct {
	Number int      `json:"number"`
	Head   string   `json:"head"`
	Body   []string `json:"body"`
	Text   string   `json:"text"`
}

type grammarJSON struct {
	Start        string           `json:"start"`
	Productions  []productionJSON `json:"productions"`
	Grouped      []string         `json:"grouped"`
	Terminals    []string         `json:"terminals"`
	NonTerminals []string         `json:"nonTerminals"`
}

type setsRowJSON struct {
	NonTerminal string   `json:"nonTerminal"`
	First       []string `json:"first"`
	Follow      []string `json:"follow"`
}

type tableCellJSON struct {
	Column string `json:"column"`
	Number int    `json:"number"` // 0 means an empty (error) cell
	Text   string `json:"text"`
}

type tableRowJSON struct {
	NonTerminal string          `json:"nonTerminal"`
	Cells       []tableCellJSON `json:"cells"`
}

type tableJSON struct {
	Columns   []string       `json:"columns"`
	Rows      []tableRowJSON `json:"rows"`
	Conflicts []string       `json:"conflicts"`
}

type cellRefJSON struct {
	NonTerminal string `json:"nonTerminal"`
	Column      string `json:"column"`
}

type stepJSON struct {
	Number int          `json:"number"`
	Stack  []string     `json:"stack"`
	Input  []string     `json:"input"`
	Action string       `json:"action"`
	Cell   *cellRefJSON `json:"cell"` // table cell consulted by this step, if any
}

type tokenJSON struct {
	Kind   string `json:"kind"`
	Lexeme string `json:"lexeme"`
	Pos    int    `json:"pos"`
}

type resultJSON struct {
	Accepted bool        `json:"accepted"`
	Message  string      `json:"message"`
	Tokens   []tokenJSON `json:"tokens"`
	Steps    []stepJSON  `json:"steps"`
}

var (
	grammar = ll1.Default()
	table   = ll1.BuildTable(grammar)
	rng     = rand.New(rand.NewSource(time.Now().UnixNano()))
)

func toJSON(value any) any {
	encoded, err := json.Marshal(value)
	if err != nil {
		return `{"error":"` + err.Error() + `"}`
	}
	return string(encoded)
}

func grammarValue() any {
	out := grammarJSON{
		Start:        grammar.Start,
		Grouped:      strings.Split(strings.TrimSpace(grammar.String()), "\n"),
		Terminals:    grammar.Terminals,
		NonTerminals: grammar.NonTerminals,
	}
	for _, production := range grammar.Productions {
		out.Productions = append(out.Productions, productionJSON{
			Number: production.Number,
			Head:   production.Head,
			Body:   production.Body,
			Text:   production.String(),
		})
	}
	return toJSON(out)
}

func setsValue() any {
	first, follow := grammar.First(), grammar.Follow()
	rows := []setsRowJSON{}
	for _, nonTerminal := range grammar.NonTerminals {
		rows = append(rows, setsRowJSON{nonTerminal, first[nonTerminal], follow[nonTerminal]})
	}
	return toJSON(rows)
}

func tableValue() any {
	out := tableJSON{Columns: table.Columns, Conflicts: append([]string{}, table.Conflicts...), Rows: []tableRowJSON{}}
	for _, nonTerminal := range grammar.NonTerminals {
		row := tableRowJSON{NonTerminal: nonTerminal}
		for _, column := range table.Columns {
			cell := tableCellJSON{Column: column}
			if production, ok := table.Production(nonTerminal, column); ok {
				cell.Number, cell.Text = production.Number, production.String()
			}
			row.Cells = append(row.Cells, cell)
		}
		out.Rows = append(out.Rows, row)
	}
	return toJSON(out)
}

// consultedCell returns the table cell a step reads: the stack top against the
// current input token, when the top is a non-terminal with an entry there.
func consultedCell(step ll1.Step) *cellRefJSON {
	if len(step.Stack) == 0 || len(step.Input) == 0 {
		return nil
	}
	top := step.Stack[len(step.Stack)-1]
	if !grammar.IsNonTerminal(top) {
		return nil
	}
	if _, ok := table.Production(top, step.Input[0]); !ok {
		return nil
	}
	return &cellRefJSON{NonTerminal: top, Column: step.Input[0]}
}

func parseValue(input string) any {
	result := ll1.Parse(table, input)
	out := resultJSON{Accepted: result.Accepted, Message: result.Message, Tokens: []tokenJSON{}, Steps: []stepJSON{}}
	for _, token := range result.Tokens {
		out.Tokens = append(out.Tokens, tokenJSON{token.Kind, token.Lexeme, token.Pos})
	}
	for _, step := range result.Steps {
		out.Steps = append(out.Steps, stepJSON{step.Number, step.Stack, step.Input, step.Action, consultedCell(step)})
	}
	return toJSON(out)
}

func main() {
	api := map[string]any{
		"grammar": js.FuncOf(func(js.Value, []js.Value) any { return grammarValue() }),
		"sets":    js.FuncOf(func(js.Value, []js.Value) any { return setsValue() }),
		"table":   js.FuncOf(func(js.Value, []js.Value) any { return tableValue() }),
		"parse": js.FuncOf(func(_ js.Value, args []js.Value) any {
			input := ""
			if len(args) > 0 {
				input = args[0].String()
			}
			return parseValue(input)
		}),
		"generate": js.FuncOf(func(js.Value, []js.Value) any {
			return ll1.Generate(grammar, rng, maxGenerateDepth)
		}),
		"generateInvalid": js.FuncOf(func(js.Value, []js.Value) any {
			return ll1.GenerateInvalid(grammar, rng)
		}),
	}
	js.Global().Set("ll1", js.ValueOf(api))
	select {}
}
