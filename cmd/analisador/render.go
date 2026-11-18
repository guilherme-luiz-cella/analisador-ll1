package main

import (
	"strings"
	"unicode/utf8"
)

// Estilos ANSI; só são escritos quando a paleta está ligada.
const (
	styleReset = "\033[0m"
	styleBold  = "\033[1m"
	styleTitle = "\033[1;36m"
	styleGood  = "\033[1;32m"
	styleBad   = "\033[1;31m"
	styleDim   = "\033[2m"
)

// palette envolve o texto em estilos ANSI, ou o deixa puro quando desligada.
type palette struct {
	enabled bool
}

func (p palette) paint(style, text string) string {
	if !p.enabled || text == "" {
		return text
	}
	return style + text + styleReset
}

// box desenha uma moldura de linha dupla em volta das linhas.
func box(p palette, lines ...string) string {
	width := 0
	for _, line := range lines {
		width = max(width, utf8.RuneCountInString(line))
	}
	var sb strings.Builder
	sb.WriteString("╔" + strings.Repeat("═", width+4) + "╗\n")
	for _, line := range lines {
		sb.WriteString("║  " + p.paint(styleTitle, pad(line, width)) + "  ║\n")
	}
	sb.WriteString("╚" + strings.Repeat("═", width+4) + "╝\n")
	return sb.String()
}

// cellStyle escolhe o estilo de uma célula do corpo; "" é texto puro.
type cellStyle func(column int, text string) string

// table desenha uma grade alinhada com linha de cabeçalho. O estilo entra
// depois do preenchimento com espaços, para os códigos ANSI não contarem na
// largura e não desalinharem as colunas.
func table(p palette, headers []string, rows [][]string, style cellStyle) string {
	widths := make([]int, len(headers))
	for i, h := range headers {
		widths[i] = utf8.RuneCountInString(h)
	}
	for _, row := range rows {
		for i, cell := range row {
			widths[i] = max(widths[i], utf8.RuneCountInString(cell))
		}
	}
	border := func(left, middle, right string) string {
		parts := make([]string, len(widths))
		for i, w := range widths {
			parts[i] = strings.Repeat("─", w+2)
		}
		return left + strings.Join(parts, middle) + right + "\n"
	}
	line := func(cells []string, paint func(column int, text string) string) string {
		parts := make([]string, len(widths))
		for i, w := range widths {
			parts[i] = " " + paint(i, pad(cells[i], w)) + " "
		}
		return "│" + strings.Join(parts, "│") + "│\n"
	}

	var sb strings.Builder
	sb.WriteString(border("┌", "┬", "┐"))
	sb.WriteString(line(headers, func(_ int, text string) string { return p.paint(styleBold, text) }))
	sb.WriteString(border("├", "┼", "┤"))
	for _, row := range rows {
		sb.WriteString(line(row, func(column int, text string) string {
			if style == nil {
				return text
			}
			return p.paint(style(column, row[column]), text)
		}))
	}
	sb.WriteString(border("└", "┴", "┘"))
	return sb.String()
}

func pad(text string, width int) string {
	return text + strings.Repeat(" ", width-utf8.RuneCountInString(text))
}
