package ll1

import (
	"math"
	"math/rand"
	"strconv"
	"strings"
)

// identifierLexemes são nomes curtos para caber na tela do projetor.
var identifierLexemes = []string{"x", "y", "z", "a", "b", "soma", "total"}

// maxNumberLexeme mantém os números gerados com até dois dígitos, fáceis de ler.
const maxNumberLexeme = 99

// DefaultMaxDepth deixa as sentenças geradas curtas o bastante para ler o
// rastro no projetor; o terminal, a web e GenerateInvalid usam este valor.
const DefaultMaxDepth = 4

// illegalCharacters provocam erro léxico; os que forem terminais da gramática
// atual são pulados.
var illegalCharacters = []string{"#", "@", "!", "?", "&", "%", ";"}

// Generate monta uma sentença válida aleatória por derivação mais à esquerda
// a partir do símbolo inicial. Abaixo de maxDepth toda produção tem a mesma
// chance; a partir dele só entram as que terminam a derivação mais rápido.
// Os tokens saem separados por espaço: "x + 3 * ( y )". Supõe que todo não
// terminal derive alguma sentença (vale para todas as gramáticas prontas).
func Generate(g *Grammar, rng *rand.Rand, maxDepth int) string {
	heights := g.productionHeights()
	var lexemes []string
	var expand func(symbol string, depth int)
	expand = func(symbol string, depth int) {
		if !g.IsNonTerminal(symbol) {
			lexemes = append(lexemes, lexemeFor(symbol, rng))
			return
		}
		p := chooseProduction(g.ProductionsOf(symbol), heights, rng, depth >= maxDepth)
		if p.IsEmpty() {
			return
		}
		for _, child := range p.Body {
			expand(child, depth+1)
		}
	}
	expand(g.Start, 0)
	return strings.Join(lexemes, " ")
}

// GenerateInvalid estraga uma sentença válida (tira um token, duplica um
// operador, põe um parêntese sem par ou um caractere ilegal) e tenta de novo
// até o analisador rejeitar o resultado.
func GenerateInvalid(g *Grammar, rng *rand.Rand) string {
	table := BuildTable(g)
	for {
		lexemes := strings.Fields(Generate(g, rng, DefaultMaxDepth))
		candidate := strings.Join(mutate(g, lexemes, rng), " ")
		// Tirar o único token de "x" deixa a entrada vazia: é rejeitada, mas
		// não mostra nada na tela, então tenta de novo.
		if strings.TrimSpace(candidate) == "" {
			continue
		}
		if !Parse(table, candidate).Accepted {
			return candidate
		}
	}
}

// mutate aplica um estrago aleatório numa cópia dos lexemas. Os "operadores"
// são os terminais de um caractere da gramática. Um estrago que não se aplica
// devolve os lexemas sem mudança, e GenerateInvalid tenta de novo.
func mutate(g *Grammar, lexemes []string, rng *rand.Rand) []string {
	operators := g.punctuation()
	const (
		dropToken = iota
		duplicateOperator
		unbalancedParenthesis
		illegalCharacter
		mutationCount
	)
	switch rng.Intn(mutationCount) {
	case dropToken:
		if len(lexemes) == 0 {
			return lexemes
		}
		i := rng.Intn(len(lexemes))
		return append(append([]string{}, lexemes[:i]...), lexemes[i+1:]...)
	case duplicateOperator:
		var operatorIndexes []int
		for i, lexeme := range lexemes {
			if contains(operators, lexeme) {
				operatorIndexes = append(operatorIndexes, i)
			}
		}
		if len(operatorIndexes) > 0 {
			i := operatorIndexes[rng.Intn(len(operatorIndexes))]
			return insertAt(lexemes, i, lexemes[i])
		}
		if len(operators) == 0 {
			return lexemes
		}
		return insertAt(lexemes, rng.Intn(len(lexemes)+1), operators[0])
	case unbalancedParenthesis:
		parenthesis := []string{"(", ")"}[rng.Intn(2)]
		return insertAt(lexemes, rng.Intn(len(lexemes)+1), parenthesis)
	default:
		var illegal []string
		for _, char := range illegalCharacters {
			if !g.IsTerminal(char) {
				illegal = append(illegal, char)
			}
		}
		if len(illegal) == 0 {
			return lexemes
		}
		char := illegal[rng.Intn(len(illegal))]
		return insertAt(lexemes, rng.Intn(len(lexemes)+1), char)
	}
}

func insertAt(lexemes []string, i int, lexeme string) []string {
	result := append([]string{}, lexemes[:i]...)
	result = append(result, lexeme)
	return append(result, lexemes[i:]...)
}

func lexemeFor(terminal string, rng *rand.Rand) string {
	switch terminal {
	case TerminalID:
		return identifierLexemes[rng.Intn(len(identifierLexemes))]
	case TerminalNum:
		return strconv.Itoa(rng.Intn(maxNumberLexeme + 1))
	default:
		return terminal
	}
}

// chooseProduction sorteia entre todas as opções ou, com shortest, só entre
// as que terminam a derivação mais rápido, para a recursão acabar.
func chooseProduction(options []Production, heights map[int]int, rng *rand.Rand, shortest bool) Production {
	if shortest {
		best := math.MaxInt
		var fastest []Production
		for _, p := range options {
			switch h := heights[p.Number]; {
			case h < best:
				best = h
				fastest = []Production{p}
			case h == best:
				fastest = append(fastest, p)
			}
		}
		options = fastest
	}
	return options[rng.Intn(len(options))]
}

// productionHeights dá, por número de produção, a altura da menor árvore de
// derivação que começa por ela. Ponto fixo: a altura de um não terminal só
// cai quando a de algum símbolo do corpo cai, então repete até parar.
func (g *Grammar) productionHeights() map[int]int {
	symbolHeight := map[string]int{}
	for _, nt := range g.NonTerminals {
		symbolHeight[nt] = math.MaxInt
	}
	heightOf := func(p Production) int {
		height := 0
		for _, symbol := range p.Body {
			if h, isNonTerminal := symbolHeight[symbol]; isNonTerminal {
				if h == math.MaxInt {
					return math.MaxInt
				}
				height = max(height, h)
			}
		}
		return height + 1
	}
	for changed := true; changed; {
		changed = false
		for _, p := range g.Productions {
			if h := heightOf(p); h < symbolHeight[p.Head] {
				symbolHeight[p.Head] = h
				changed = true
			}
		}
	}
	heights := make(map[int]int, len(g.Productions))
	for _, p := range g.Productions {
		heights[p.Number] = heightOf(p)
	}
	return heights
}
