// Command analisador é a versão de terminal do analisador LL(1) preditivo
// tabular: mostra a gramática, os conjuntos FIRST e FOLLOW e a tabela de
// parsing, e rastreia o reconhecimento de sentenças digitadas ou geradas.
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

// menuOption é uma entrada do menu principal.
type menuOption struct {
	key    string
	label  string
	action func(a *app) bool // false quando a entrada acabou
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

// app guarda o estado do analisador e a entrada e saída de uma sessão.
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

// run mostra o menu até o usuário escolher 0 ou a entrada acabar.
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

// use torna preset a gramática atual; a tabela é montada uma vez aqui.
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
			cell := "" // célula vazia é erro sintático
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
	a.analyze(ll1.Generate(a.grammar, a.rng, ll1.DefaultMaxDepth))
	return true
}

func (a *app) generateInvalid() bool {
	a.section("GERAR SENTENÇA INVÁLIDA")
	a.analyze(ll1.GenerateInvalid(a.grammar, a.rng))
	return true
}

// choosePreset lista as gramáticas prontas por grupo, com autor e fonte, e
// troca para a escolhida pelo número.
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

// sourceText escreve "Fonte: autor, título", mais o aviso de adaptação para LL(1).
func sourceText(preset ll1.Preset) string {
	text := "Fonte: " + preset.Author + ", " + preset.Source
	if preset.Adapted {
		text += " (adaptado para LL(1))"
	}
	return text
}

// analyze analisa a sentença e imprime os tokens, o rastro e o resultado.
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
		func(column int, text string) string {
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

// readLine mostra o prompt e devolve a próxima linha, ou false no fim da entrada.
func (a *app) readLine(prompt string) (string, bool) {
	a.print(prompt)
	if !a.in.Scan() {
		return "", false
	}
	// Mantém os espaços do início e do meio para as posições dos tokens
	// baterem com o que foi digitado.
	return strings.TrimRight(a.in.Text(), "\r"), true
}

func (a *app) print(text string) {
	fmt.Fprint(a.out, text)
}

func setText(symbols []string) string {
	return "{ " + strings.Join(symbols, ", ") + " }"
}
