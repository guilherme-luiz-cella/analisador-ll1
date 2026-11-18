package ll1

import (
	"math"
	"math/rand"
	"strconv"
	"strings"
)

// identifierLexemes are the names used for generated "id" tokens.
var identifierLexemes = []string{"x", "y", "z", "a", "b", "soma", "total"}

// maxNumberLexeme is the largest value used for generated "num" tokens.
const maxNumberLexeme = 99

// invalidBaseDepth is the depth of the valid sentence GenerateInvalid mutates.
const invalidBaseDepth = 4

// illegalCharacters are inserted to cause lexical errors.
var illegalCharacters = []string{"#", "@", "!", "?", "&", "%", ";"}

// Generate builds a random valid sentence by a leftmost derivation from the
// start symbol. Below maxDepth every production is equally likely; from
// maxDepth on, only the productions that finish the derivation fastest are
// chosen. Tokens are joined by spaces: "x + 3 * ( y )".
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

// GenerateInvalid mutates a valid sentence (drops a token, duplicates an
// operator, adds an unbalanced parenthesis or inserts an illegal character)
// and retries until the parser rejects the result.
func GenerateInvalid(g *Grammar, rng *rand.Rand) string {
	table := BuildTable(g)
	for {
		lexemes := strings.Fields(Generate(g, rng, invalidBaseDepth))
		candidate := strings.Join(mutate(lexemes, rng), " ")
		if !Parse(table, candidate).Accepted {
			return candidate
		}
	}
}

// mutate applies one random kind of damage to a copy of the lexemes.
func mutate(lexemes []string, rng *rand.Rand) []string {
	const (
		dropToken = iota
		duplicateOperator
		unbalancedParenthesis
		illegalCharacter
		mutationCount
	)
	switch rng.Intn(mutationCount) {
	case dropToken:
		i := rng.Intn(len(lexemes))
		return append(append([]string{}, lexemes[:i]...), lexemes[i+1:]...)
	case duplicateOperator:
		var operatorIndexes []int
		for i, lexeme := range lexemes {
			if strings.Contains(operators, lexeme) {
				operatorIndexes = append(operatorIndexes, i)
			}
		}
		if len(operatorIndexes) > 0 {
			i := operatorIndexes[rng.Intn(len(operatorIndexes))]
			return insertAt(lexemes, i, lexemes[i])
		}
		return insertAt(lexemes, rng.Intn(len(lexemes)+1), "+")
	case unbalancedParenthesis:
		parenthesis := []string{"(", ")"}[rng.Intn(2)]
		return insertAt(lexemes, rng.Intn(len(lexemes)+1), parenthesis)
	default:
		char := illegalCharacters[rng.Intn(len(illegalCharacters))]
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

// chooseProduction picks uniformly, or among the shortest-finishing
// productions when shortest is set.
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

// productionHeights gives, per production number, the height of the
// shortest derivation tree that starts with it (fixed-point computation).
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
