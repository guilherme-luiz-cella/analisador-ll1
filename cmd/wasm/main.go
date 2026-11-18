//go:build js && wasm

// Command wasm expõe o pacote ll1 ao navegador como globalThis.ll1. Toda
// função devolve um texto JSON, menos generate e generateInvalid, que devolvem
// a própria sentença. Toda função menos presets recebe, por último, o id
// opcional de uma gramática pronta; id ausente ou desconhecido quer dizer a
// gramática do trabalho.
package main

import (
	"encoding/json"
	"math/rand"
	"syscall/js"
	"time"

	"analisador-ll1/ll1"
)

// Os tipos *JSON abaixo são o formato que web/app.js lê: os nomes dos campos
// em JSON são o contrato com a página.

type productionJSON struct {
	Number int      `json:"number"`
	Head   string   `json:"head"`
	Body   []string `json:"body"`
	First  []string `json:"first"` // FIRST do corpo, usado no painel da condição LL(1)
}

type grammarJSON struct {
	Start        string           `json:"start"`
	Productions  []productionJSON `json:"productions"`
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
	Number    int    `json:"number"`              // 0 é célula vazia (erro)
	Conflicts []int  `json:"conflicts,omitempty"` // todas as produções que disputam a célula
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
	Cell   *cellRefJSON `json:"cell"` // célula da tabela consultada neste passo, se houver
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
	Group           string   `json:"group"` // título do grupo na lista de escolha
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

// loaded é uma gramática pronta com a sua tabela, montada uma vez no início.
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

// selected acha a gramática pelo id opcional em args[index].
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

// toJSON codifica o valor. Os tipos acima só têm textos, números e listas,
// então Marshal não falha; se falhar, é erro do programa e deve aparecer.
func toJSON(value any) any {
	encoded, err := json.Marshal(value)
	if err != nil {
		panic(err)
	}
	return string(encoded)
}

func grammarValue(grammar *ll1.Grammar) any {
	out := grammarJSON{
		Start:        grammar.Start,
		Terminals:    grammar.Terminals,
		NonTerminals: grammar.NonTerminals,
	}
	for _, production := range grammar.Productions {
		out.Productions = append(out.Productions, productionJSON{
			Number: production.Number,
			Head:   production.Head,
			Body:   production.Body,
			First:  grammar.FirstOfSequence(production.Body),
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
	conflicts := map[[2]string][]int{} // [não terminal, coluna] -> produções em disputa
	for _, conflict := range table.ConflictCells {
		conflicts[[2]string{conflict.NonTerminal, conflict.Column}] = conflict.Numbers
	}
	for _, nonTerminal := range grammar.NonTerminals {
		row := tableRowJSON{NonTerminal: nonTerminal}
		for _, column := range table.Columns {
			row.Cells = append(row.Cells, tableCellJSON{
				Column:    column,
				Number:    table.Cells[nonTerminal][column], // 0 quando a célula está vazia
				Conflicts: conflicts[[2]string{nonTerminal, column}],
			})
		}
		out.Rows = append(out.Rows, row)
	}
	return toJSON(out)
}

// consultedCell devolve a célula que o passo lê: o topo da pilha contra o
// token atual, quando o topo é não terminal e a célula tem produção. A página
// usa isso para destacar a célula e montar a derivação.
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
			return ll1.Generate(selected(args, 0).preset.Grammar, rng, ll1.DefaultMaxDepth)
		}),
		"generateInvalid": js.FuncOf(func(_ js.Value, args []js.Value) any {
			return ll1.GenerateInvalid(selected(args, 0).preset.Grammar, rng)
		}),
	}
	// select {} mantém o programa vivo: as funções acima só respondem enquanto
	// main não termina.
	js.Global().Set("ll1", js.ValueOf(api))
	select {}
}
