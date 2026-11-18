package ll1

// symbolSet é um conjunto de símbolos sem ordem, usado durante as repetições;
// a ordem de exibição só é aplicada no fim, por ordered.
type symbolSet map[string]bool

// noSkip é o valor de skip em addAll quando nenhum símbolo deve ficar de fora.
const noSkip = ""

// addAll junta todos os símbolos de other, menos skip, e diz se s cresceu.
// É esse "cresceu" que decide se o ponto fixo precisa de mais uma volta.
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

// First calcula o FIRST de cada não terminal pelo método do ponto fixo.
// Cada conjunto sai na ordem de Terminals, com ε no fim quando presente.
func (g *Grammar) First() map[string][]string {
	first := g.firstSets()
	result := make(map[string][]string, len(first))
	for nt, set := range first {
		result[nt] = g.ordered(set, Epsilon)
	}
	return result
}

// FirstOfSequence devolve o FIRST de uma sequência de símbolos, na ordem de
// Terminals com ε no fim. Uma sequência vazia (ou só ε) dá {ε}.
func (g *Grammar) FirstOfSequence(seq []string) []string {
	return g.ordered(g.firstOfSequence(g.firstSets(), seq), Epsilon)
}

// Follow calcula o FOLLOW de cada não terminal pelo método do ponto fixo.
// Cada conjunto sai na ordem de Terminals, com $ no fim quando presente.
func (g *Grammar) Follow() map[string][]string {
	follow := g.followSets(g.firstSets())
	result := make(map[string][]string, len(follow))
	for nt, set := range follow {
		result[nt] = g.ordered(set, EndMarker)
	}
	return result
}

// firstSets aplica as regras do FIRST até nenhum conjunto mudar.
func (g *Grammar) firstSets() map[string]symbolSet {
	first := make(map[string]symbolSet, len(g.NonTerminals))
	for _, nt := range g.NonTerminals {
		first[nt] = symbolSet{}
	}
	// Ponto fixo: um FIRST depende de outros ainda incompletos, então repete
	// as regras até nenhum conjunto crescer.
	for changed := true; changed; {
		changed = false
		for _, p := range g.Productions {
			// Para A → X1...Xn, FIRST(A) recebe FIRST(X1...Xn).
			if first[p.Head].addAll(g.firstOfSequence(first, p.Body), noSkip) {
				changed = true
			}
		}
	}
	return first
}

// firstOfSequence usa os conjuntos FIRST atuais, talvez ainda parciais.
func (g *Grammar) firstOfSequence(first map[string]symbolSet, seq []string) symbolSet {
	result := symbolSet{}
	for _, symbol := range seq {
		if symbol == Epsilon {
			continue
		}
		// FIRST de um terminal é ele mesmo, e a sequência começa por ele.
		if !g.IsNonTerminal(symbol) {
			result[symbol] = true
			return result
		}
		// Soma FIRST(Xi) sem ε; só passa para Xi+1 se Xi deriva ε.
		result.addAll(first[symbol], Epsilon)
		if !first[symbol][Epsilon] {
			return result
		}
	}
	// Todos os Xi derivam ε (ou o corpo é ε), então ε ∈ FIRST.
	result[Epsilon] = true
	return result
}

// followSets aplica as regras do FOLLOW até nenhum conjunto mudar.
func (g *Grammar) followSets(first map[string]symbolSet) map[string]symbolSet {
	follow := make(map[string]symbolSet, len(g.NonTerminals))
	for _, nt := range g.NonTerminals {
		follow[nt] = symbolSet{}
	}
	// Regra 1 do FOLLOW: $ ∈ FOLLOW(S), S o símbolo inicial.
	follow[g.Start][EndMarker] = true
	// Ponto fixo: a regra 3 copia FOLLOW de um não terminal para outro, então
	// repete até nada mudar.
	for changed := true; changed; {
		changed = false
		for _, p := range g.Productions {
			for i, symbol := range p.Body {
				if !g.IsNonTerminal(symbol) {
					continue
				}
				// Regra 2 do FOLLOW: A → αBβ, FIRST(β) − {ε} entra em FOLLOW(B).
				rest := g.firstOfSequence(first, p.Body[i+1:])
				if follow[symbol].addAll(rest, Epsilon) {
					changed = true
				}
				// Regra 3 do FOLLOW: A → αB, ou A → αBβ com ε ∈ FIRST(β),
				// FOLLOW(A) entra em FOLLOW(B).
				if rest[Epsilon] && follow[symbol].addAll(follow[p.Head], noSkip) {
					changed = true
				}
			}
		}
	}
	return follow
}

// ordered lista o conjunto na ordem de Terminals e põe extra no fim, se estiver
// nele, para a tela mostrar sempre a mesma ordem.
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
