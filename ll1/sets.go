package ll1

// symbolSet is an unordered set of grammar symbols used while iterating.
type symbolSet map[string]bool

// addAll adds every symbol of other except skip and reports whether s grew.
func (s symbolSet) addAll(other symbolSet, skip string) bool {
	grew := false
	for symbol := range other {
		if symbol != skip && !s[symbol] {
			s[symbol] = true
			grew = true
		}
	}
	return grew
}

// First computes FIRST of every non-terminal with the fixed-point algorithm.
// Each set is ordered like Terminals, with ε last when present.
func (g *Grammar) First() map[string][]string {
	first := g.firstSets()
	result := make(map[string][]string, len(first))
	for nt, set := range first {
		result[nt] = g.ordered(set, Epsilon)
	}
	return result
}

// FirstOfSequence returns FIRST of a sequence of symbols, ordered like
// Terminals with ε last. An empty sequence (or just ε) gives {ε}.
func (g *Grammar) FirstOfSequence(seq []string) []string {
	return g.ordered(g.firstOfSequence(g.firstSets(), seq), Epsilon)
}

// Follow computes FOLLOW of every non-terminal with the fixed-point algorithm.
// Each set is ordered like Terminals, with $ last when present.
func (g *Grammar) Follow() map[string][]string {
	follow := g.followSets(g.firstSets())
	result := make(map[string][]string, len(follow))
	for nt, set := range follow {
		result[nt] = g.ordered(set, EndMarker)
	}
	return result
}

// firstSets repeats the FIRST rules until no set changes.
func (g *Grammar) firstSets() map[string]symbolSet {
	first := make(map[string]symbolSet, len(g.NonTerminals))
	for _, nt := range g.NonTerminals {
		first[nt] = symbolSet{}
	}
	for changed := true; changed; {
		changed = false
		for _, p := range g.Productions {
			if first[p.Head].addAll(g.firstOfSequence(first, p.Body), "") {
				changed = true
			}
		}
	}
	return first
}

// firstOfSequence uses the current (possibly partial) FIRST sets.
func (g *Grammar) firstOfSequence(first map[string]symbolSet, seq []string) symbolSet {
	result := symbolSet{}
	for _, symbol := range seq {
		if symbol == Epsilon {
			continue
		}
		if !g.IsNonTerminal(symbol) {
			result[symbol] = true
			return result
		}
		result.addAll(first[symbol], Epsilon)
		if !first[symbol][Epsilon] {
			return result
		}
	}
	// Every symbol can vanish (or the sequence is empty).
	result[Epsilon] = true
	return result
}

// followSets repeats the FOLLOW rules until no set changes.
func (g *Grammar) followSets(first map[string]symbolSet) map[string]symbolSet {
	follow := make(map[string]symbolSet, len(g.NonTerminals))
	for _, nt := range g.NonTerminals {
		follow[nt] = symbolSet{}
	}
	follow[g.Start][EndMarker] = true
	for changed := true; changed; {
		changed = false
		for _, p := range g.Productions {
			for i, symbol := range p.Body {
				if !g.IsNonTerminal(symbol) {
					continue
				}
				rest := g.firstOfSequence(first, p.Body[i+1:])
				if follow[symbol].addAll(rest, Epsilon) {
					changed = true
				}
				if rest[Epsilon] && follow[symbol].addAll(follow[p.Head], "") {
					changed = true
				}
			}
		}
	}
	return follow
}

// ordered lists the set in Terminals order, then the extra symbol if present.
func (g *Grammar) ordered(set symbolSet, extra string) []string {
	result := []string{}
	for _, t := range g.Terminals {
		if set[t] {
			result = append(result, t)
		}
	}
	if set[extra] {
		result = append(result, extra)
	}
	return result
}
