//go:build js && wasm

// Command wasm exposes the ll1 package to the browser as globalThis.ll1.
// Every function returns a JSON string, except generate and generateInvalid,
// which return the sentence itself. Every function except presets takes an
// optional preset id as its last argument; a missing or unknown id means the
// assignment grammar.
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
	Column    string `json:"column"`
	Number    int    `json:"number"` // 0 means an empty (error) cell
	Text      string `json:"text"`
	Conflicts []int  `json:"conflicts,omitempty"` // every production claiming a conflicting cell
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

type presetJSON struct {
	ID              string   `json:"id"`
	Name            string   `json:"name"`
	Description     string   `json:"description"`
	Group           string   `json:"group"` // pt-BR label of the dropdown group
	IsDefault       bool     `json:"isDefault"`
	Author          string   `json:"author"`
	Source          string   `json:"source"`
	URL             string   `json:"url"`
	Adapted         bool     `json:"adapted"`
	IsLL1           bool     `json:"isLL1"`
	Terminals       []string `json:"terminals"`
	Examples        []string `json:"examples"`
	InvalidExamples []string `json:"invalidExamples"`
}

// loaded is one preset with its table, built once at start.
type loaded struct {
	preset ll1.Preset
	table  *ll1.Table
}

var (
	presets   = loadPresets()
	presetMap = indexPresets(presets)
	rng       = rand.New(rand.NewSource(time.Now().UnixNano()))
)

func loadPresets() []loaded {
	var out []loaded
	for _, preset := range ll1.Presets() {
		out = append(out, loaded{preset, ll1.BuildTable(preset.Grammar)})
	}
	return out
}

func indexPresets(list []loaded) map[string]loaded {
	out := make(map[string]loaded, len(list))
	for _, item := range list {
		out[item.preset.ID] = item
	}
	return out
}

// selected resolves the optional preset id at args[index].
func selected(args []js.Value, index int) loaded {
	if len(args) > index && args[index].Type() == js.TypeString {
		if item, ok := presetMap[args[index].String()]; ok {
			return item
		}
	}
	return presetMap[ll1.DefaultPresetID]
}

func presetsValue() any {
	out := []presetJSON{}
	for _, item := range presets {
		p := item.preset
		out = append(out, presetJSON{
			ID: p.ID, Name: p.Name, Description: p.Description, Group: p.Group.Label(),
			IsDefault: p.ID == ll1.DefaultPresetID, Author: p.Author, Source: p.Source, URL: p.URL,
			Adapted: p.Adapted, IsLL1: len(item.table.Conflicts) == 0, Terminals: p.Grammar.Terminals,
			Examples: p.Examples, InvalidExamples: p.InvalidExamples,
		})
	}
	return toJSON(out)
}

func toJSON(value any) any {
	encoded, err := json.Marshal(value)
	if err != nil {
		return `{"error":"` + err.Error() + `"}`
	}
	return string(encoded)
}

func grammarValue(grammar *ll1.Grammar) any {
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

func setsValue(grammar *ll1.Grammar) any {
	first, follow := grammar.First(), grammar.Follow()
	rows := []setsRowJSON{}
	for _, nonTerminal := range grammar.NonTerminals {
		rows = append(rows, setsRowJSON{nonTerminal, first[nonTerminal], follow[nonTerminal]})
	}
	return toJSON(rows)
}

func tableValue(table *ll1.Table) any {
	grammar := table.Grammar
	out := tableJSON{Columns: table.Columns, Conflicts: append([]string{}, table.Conflicts...), Rows: []tableRowJSON{}}
	for _, nonTerminal := range grammar.NonTerminals {
		row := tableRowJSON{NonTerminal: nonTerminal}
		for _, column := range table.Columns {
			cell := tableCellJSON{Column: column}
			if production, ok := table.Production(nonTerminal, column); ok {
				cell.Number, cell.Text = production.Number, production.String()
			}
			for _, conflict := range table.ConflictCells {
				if conflict.NonTerminal == nonTerminal && conflict.Column == column {
					cell.Conflicts = conflict.Numbers
				}
			}
			row.Cells = append(row.Cells, cell)
		}
		out.Rows = append(out.Rows, row)
	}
	return toJSON(out)
}

// consultedCell returns the table cell a step reads: the stack top against the
// current input token, when the top is a non-terminal with an entry there.
func consultedCell(table *ll1.Table, step ll1.Step) *cellRefJSON {
	if len(step.Stack) == 0 || len(step.Input) == 0 {
		return nil
	}
	top := step.Stack[len(step.Stack)-1]
	if !table.Grammar.IsNonTerminal(top) {
		return nil
	}
	if _, ok := table.Production(top, step.Input[0]); !ok {
		return nil
	}
	return &cellRefJSON{NonTerminal: top, Column: step.Input[0]}
}

func parseValue(table *ll1.Table, input string) any {
	result := ll1.Parse(table, input)
	out := resultJSON{Accepted: result.Accepted, Message: result.Message, Tokens: []tokenJSON{}, Steps: []stepJSON{}}
	for _, token := range result.Tokens {
		out.Tokens = append(out.Tokens, tokenJSON{token.Kind, token.Lexeme, token.Pos})
	}
	for _, step := range result.Steps {
		out.Steps = append(out.Steps, stepJSON{step.Number, step.Stack, step.Input, step.Action, consultedCell(table, step)})
	}
	return toJSON(out)
}

func main() {
	api := map[string]any{
		"presets": js.FuncOf(func(js.Value, []js.Value) any { return presetsValue() }),
		"grammar": js.FuncOf(func(_ js.Value, args []js.Value) any { return grammarValue(selected(args, 0).preset.Grammar) }),
		"sets":    js.FuncOf(func(_ js.Value, args []js.Value) any { return setsValue(selected(args, 0).preset.Grammar) }),
		"table":   js.FuncOf(func(_ js.Value, args []js.Value) any { return tableValue(selected(args, 0).table) }),
		"parse": js.FuncOf(func(_ js.Value, args []js.Value) any {
			input := ""
			if len(args) > 0 {
				input = args[0].String()
			}
			return parseValue(selected(args, 1).table, input)
		}),
		"generate": js.FuncOf(func(_ js.Value, args []js.Value) any {
			return ll1.Generate(selected(args, 0).preset.Grammar, rng, maxGenerateDepth)
		}),
		"generateInvalid": js.FuncOf(func(_ js.Value, args []js.Value) any {
			return ll1.GenerateInvalid(selected(args, 0).preset.Grammar, rng)
		}),
	}
	js.Global().Set("ll1", js.ValueOf(api))
	select {}
}
