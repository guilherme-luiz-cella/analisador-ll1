// Command analisador is the terminal front end of the LL(1) predictive
// table-driven parser: it shows the grammar, the FIRST and FOLLOW sets and the
// parsing table, and traces the recognition of typed or generated sentences.
package main

import (
	"bufio"
	"fmt"
	"io"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	"analisador-ll1/ll1"
)

// generatedMaxDepth keeps generated sentences short enough to trace on screen.
const generatedMaxDepth = 4

// emptyCell marks an error entry of the parsing table.
const emptyCell = ""

// menuOption is one entry of the main menu.
type menuOption struct {
	key    string
	label  string
	action func(a *app) bool // false when input ended
}

var menu = []menuOption{
	{"1", "Gramática", (*app).showGrammar},
	{"2", "Conjuntos FIRST e FOLLOW", (*app).showSets},
	{"3", "Tabela de parsing", (*app).showTable},
	{"4", "Digitar sentença", (*app).typeSentence},
	{"5", "Gerar sentença válida", (*app).generateValid},
	{"6", "Gerar sentença inválida", (*app).generateInvalid},
	{"7", "Trocar gramática", (*app).choosePreset},
}

const exitKey = "0"

// app holds the parser state and the terminal streams of one session.
type app struct {
	preset  ll1.Preset
	grammar *ll1.Grammar
	table   *ll1.Table
	rng     *rand.Rand
	in      *bufio.Scanner
	out     io.Writer
	colors  palette
}

func main() {
	canColor, restore := prepareConsole()
	defer restore()
	colors := os.Getenv("NO_COLOR") == "" && canColor
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	run(os.Stdin, os.Stdout, colors, rng)
}

// run shows the menu until the user picks 0 or the input ends.
func run(in io.Reader, out io.Writer, colors bool, rng *rand.Rand) {
	a := &app{
		rng:    rng,
		in:     bufio.NewScanner(in),
		out:    out,
		colors: palette{enabled: colors},
	}
	preset, _ := ll1.PresetByID(ll1.DefaultPresetID)
	a.use(preset)
	a.print(box(a.colors,
		"ANALISADOR SINTÁTICO LL(1)",
		"Preditivo tabular (não recursivo)"))
	for {
		a.printMenu()
		choice, ok := a.readLine("Escolha uma opção: ")
		choice = strings.TrimSpace(choice)
		if !ok || choice == exitKey {
			a.print("\nAté logo!\n")
			return
		}
		option, found := findOption(choice)
		if !found {
			a.print(a.colors.paint(styleBad, "Opção inválida.") + "\n")
			continue
		}
		if !option.action(a) || !a.waitForEnter() {
			a.print("\nAté logo!\n")
			return
		}
	}
}

func findOption(key string) (menuOption, bool) {
	for _, option := range menu {
		if option.key == key {
			return option, true
		}
	}
	return menuOption{}, false
}

// use makes preset the current grammar.
func (a *app) use(preset ll1.Preset) {
	a.preset = preset
	a.grammar = preset.Grammar
	a.table = ll1.BuildTable(preset.Grammar)
}

func (a *app) printMenu() {
	a.print("\n" + a.colors.paint(styleTitle, "MENU") + "  " + a.colors.paint(styleDim, "Gramática: "+a.preset.Name) + "\n")
	for _, option := range menu {
		a.print(fmt.Sprintf("  %s  %s\n", a.colors.paint(styleBold, option.key), option.label))
	}
	a.print(fmt.Sprintf("  %s  %s\n\n", a.colors.paint(styleBold, exitKey), "Sair"))
}

func (a *app) showGrammar() bool {
	a.section("GRAMÁTICA")
	a.print(a.colors.paint(styleBold, a.preset.Name) + "\n" + a.preset.Description + "\n")
	if a.preset.Author != "" {
		a.print(sourceText(a.preset) + "\n")
	}
	a.print("\n" + a.grammar.String())
	a.print(fmt.Sprintf("\nSímbolo inicial: %s\n", a.grammar.Start))
	a.print(fmt.Sprintf("Não terminais:   %s\n", strings.Join(a.grammar.NonTerminals, " ")))
	a.print(fmt.Sprintf("Terminais:       %s (fim da entrada: %s)\n\n", strings.Join(a.grammar.Terminals, " "), ll1.EndMarker))
	a.print("Produções numeradas (os números aparecem na tabela de parsing):\n")
	a.print(a.productionList())
	return true
}

func (a *app) showSets() bool {
	a.section("CONJUNTOS FIRST E FOLLOW")
	first, follow := a.grammar.First(), a.grammar.Follow()
	var rows [][]string
	for _, nt := range a.grammar.NonTerminals {
		rows = append(rows, []string{nt, setText(first[nt]), setText(follow[nt])})
	}
	a.print(table(a.colors, []string{"Não terminal", "FIRST", "FOLLOW"}, rows, nil))
	return true
}

func (a *app) showTable() bool {
	a.section("TABELA DE PARSING LL(1)")
	headers := append([]string{"M"}, a.table.Columns...)
	var rows [][]string
	for _, nt := range a.grammar.NonTerminals {
		row := []string{nt}
		for _, column := range a.table.Columns {
			cell := emptyCell
			if number, ok := a.table.Cells[nt][column]; ok {
				cell = strconv.Itoa(number)
			}
			row = append(row, cell)
		}
		rows = append(rows, row)
	}
	a.print(table(a.colors, headers, rows, nil))
	a.print("Células vazias indicam erro sintático.\n")
	if len(a.table.Conflicts) == 0 {
		a.print("Nenhum conflito: a gramática é LL(1).\n\n")
	} else {
		a.print(a.colors.paint(styleBad, strings.Join(a.table.Conflicts, "\n")) + "\n\n")
	}
	a.print(a.productionList())
	return true
}

func (a *app) typeSentence() bool {
	a.section("DIGITAR SENTENÇA")
	sentence, ok := a.readLine("Digite a sentença: ")
	if !ok {
		return false
	}
	a.analyze(sentence)
	return true
}

func (a *app) generateValid() bool {
	a.section("GERAR SENTENÇA VÁLIDA")
	a.analyze(ll1.Generate(a.grammar, a.rng, generatedMaxDepth))
	return true
}

func (a *app) generateInvalid() bool {
	a.section("GERAR SENTENÇA INVÁLIDA")
	a.analyze(ll1.GenerateInvalid(a.grammar, a.rng))
	return true
}

// choosePreset lists the preset grammars by group, with author and source,
// and switches to the one picked by number.
func (a *app) choosePreset() bool {
	a.section("TROCAR GRAMÁTICA")
	presets := ll1.Presets()
	for i, preset := range presets {
		if i == 0 || preset.Group != presets[i-1].Group {
			a.print("\n" + a.colors.paint(styleTitle, preset.Group.Label()) + "\n")
		}
		marker := " "
		if preset.ID == a.preset.ID {
			marker = "*"
		}
		a.print(fmt.Sprintf(" %s%3d  %s\n", marker, i+1, preset.Name))
		if preset.Author != "" {
			a.print(a.colors.paint(styleDim, "        "+sourceText(preset)) + "\n")
			a.print(a.colors.paint(styleDim, "        "+preset.URL) + "\n")
		}
	}
	choice, ok := a.readLine(fmt.Sprintf("\nNúmero da gramática (1 a %d, Enter mantém a atual): ", len(presets)))
	if !ok {
		return false
	}
	choice = strings.TrimSpace(choice)
	if choice == "" {
		a.print("Gramática mantida: " + a.preset.Name + "\n")
		return true
	}
	number, err := strconv.Atoi(choice)
	if err != nil || number < 1 || number > len(presets) {
		a.print(a.colors.paint(styleBad, "Opção inválida.") + " Gramática mantida: " + a.preset.Name + "\n")
		return true
	}
	a.use(presets[number-1])
	a.print(a.colors.paint(styleGood, "Gramática atual: "+a.preset.Name) + "\n")
	return true
}

// sourceText reads "Fonte: autor, título", plus the LL(1) adaptation note.
func sourceText(preset ll1.Preset) string {
	text := "Fonte: " + preset.Author + ", " + preset.Source
	if preset.Adapted {
		text += " (adaptado para LL(1))"
	}
	return text
}

// analyze parses the sentence and prints its tokens, trace and result.
func (a *app) analyze(sentence string) {
	a.print(fmt.Sprintf("Sentença: %s\n\n", a.colors.paint(styleBold, sentence)))
	result := ll1.Parse(a.table, sentence)

	a.print("Tokens:\n")
	var tokenRows [][]string
	for _, token := range result.Tokens {
		tokenRows = append(tokenRows, []string{token.Kind, token.Lexeme, strconv.Itoa(token.Pos)})
	}
	if len(tokenRows) == 0 {
		a.print("  (nenhum token reconhecido)\n")
	} else {
		a.print(table(a.colors, []string{"Token", "Lexema", "Posição"}, tokenRows, nil))
	}

	a.print("\nPassos da análise (topo da pilha à direita):\n")
	const actionColumn = 3
	var stepRows [][]string
	for _, step := range result.Steps {
		stepRows = append(stepRows, []string{
			strconv.Itoa(step.Number),
			strings.Join(step.Stack, " "),
			strings.Join(step.Input, " "),
			step.Action,
		})
	}
	a.print(table(a.colors, []string{"Passo", "Pilha", "Entrada", "Ação"}, stepRows,
		func(_, column int, text string) string {
			switch {
			case column != actionColumn:
				return ""
			case text == ll1.ActionAccept:
				return styleGood
			case strings.HasPrefix(text, ll1.ErrorPrefix):
				return styleBad
			}
			return ""
		}))

	style := styleBad
	if result.Accepted {
		style = styleGood
	}
	a.print("\n" + a.colors.paint(style, result.Message) + "\n")
}

func (a *app) productionList() string {
	var sb strings.Builder
	for _, p := range a.grammar.Productions {
		sb.WriteString(fmt.Sprintf("  %2d. %s\n", p.Number, p))
	}
	return sb.String()
}

func (a *app) section(title string) {
	a.print("\n" + box(a.colors, title) + "\n")
}

func (a *app) waitForEnter() bool {
	_, ok := a.readLine("\n" + a.colors.paint(styleDim, "Pressione Enter para voltar ao menu"))
	return ok
}

// readLine prompts and returns the next input line, or false at end of input.
func (a *app) readLine(prompt string) (string, bool) {
	a.print(prompt)
	if !a.in.Scan() {
		return "", false
	}
	// Keep inner and leading spaces so token positions match what was typed.
	return strings.TrimRight(a.in.Text(), "\r"), true
}

func (a *app) print(text string) {
	fmt.Fprint(a.out, text)
}

func setText(symbols []string) string {
	return "{ " + strings.Join(symbols, ", ") + " }"
}
