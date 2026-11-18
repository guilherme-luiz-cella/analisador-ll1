package ll1

import "fmt"

// Step é uma linha do rastro da análise.
type Step struct {
	Number int
	Stack  []string // fundo primeiro, então o topo é o último; começa como ["$", "E"]
	Input  []string // tipos dos tokens que faltam, o atual primeiro, termina em "$"
	Action string   // ex.: "E -> T E'", "Casa 'id' (soma)", "Aceita", "Erro: ..."
}

// Result é o resultado da análise de uma sentença.
type Result struct {
	Accepted bool
	Message  string // mensagem final mostrada ao usuário
	Tokens   []Token
	Steps    []Step
}

// Textos fixos do rastro e da mensagem final.
const (
	ActionAccept    = "Aceita"
	ErrorPrefix     = "Erro"
	MessageAccepted = "Sentença ACEITA: a entrada foi reconhecida pela gramática."
	messageRejected = "Sentença REJEITADA: %s."
	// MessageNotLL1 explica por que Parse recusa uma tabela com conflitos.
	MessageNotLL1 = "A gramática não é LL(1): a tabela tem conflitos"
)

// Parse roda o analisador preditivo não recursivo sobre a entrada, com o
// léxico da gramática da tabela, e grava um Step para cada ação, inclusive o
// aceite ou o erro final. Uma tabela com conflitos é recusada com um único
// passo de erro, em vez de adivinhar uma produção.
func Parse(t *Table, input string) Result {
	start := []string{EndMarker, t.Grammar.Start}
	if len(t.Conflicts) > 0 {
		return refuseConflicts(t.Grammar, start, input)
	}
	tokens, err := TokenizeFor(t.Grammar, input)
	if err != nil {
		// TokenizeFor só falha com *LexicalError.
		return lexicalFailure(t.Grammar, start, input, err.(*LexicalError))
	}

	// Configuração inicial: pilha $ S, com o topo à direita, e entrada w$.
	result := Result{Tokens: tokens}
	stack := start
	position := 0
	// X é o topo da pilha e a é o token atual; cada volta faz uma ação.
	for {
		top := stack[len(stack)-1]
		current := tokens[position]
		step := Step{
			Number: len(result.Steps) + 1,
			Stack:  append([]string{}, stack...),
			Input:  kinds(tokens[position:]),
		}

		switch {
		// X = $ e a = $: pilha e entrada terminaram juntas, a sentença é aceita.
		case top == EndMarker && current.Kind == EndMarker:
			step.Action = ActionAccept
			result.Steps = append(result.Steps, step)
			result.Accepted = true
			result.Message = MessageAccepted
			return result

		// X terminal (ou $): tem de ser igual ao token atual; casou, desempilha X
		// e avança a entrada.
		case !t.Grammar.IsNonTerminal(top):
			if top != current.Kind {
				reason := fmt.Sprintf("%s, esperado %s", unexpected(current), describeSymbol(top))
				return reject(result, step, reason)
			}
			step.Action = matchAction(current)
			stack = stack[:len(stack)-1]
			position++

		// X não terminal: consulta M[X, a]; célula vazia é erro sintático.
		default:
			p, ok := t.Production(top, current.Kind)
			if !ok {
				reason := fmt.Sprintf("%s: M[%s, %s] vazia", unexpected(current), top, current.Kind)
				return reject(result, step, reason)
			}
			step.Action = p.String()
			stack = stack[:len(stack)-1]
			// Troca X por Y1...Yk empilhando de trás para frente, para Y1 ficar
			// no topo (derivação mais à esquerda); X → ε só desempilha.
			if !p.IsEmpty() {
				for i := len(p.Body) - 1; i >= 0; i-- {
					stack = append(stack, p.Body[i])
				}
			}
		}
		result.Steps = append(result.Steps, step)
	}
}

// reject fecha o rastro com o passo de erro e usa o mesmo motivo na mensagem
// final, para a tela e o rastro dizerem a mesma coisa.
func reject(result Result, step Step, reason string) Result {
	step.Action = ErrorPrefix + ": " + reason
	result.Steps = append(result.Steps, step)
	result.Message = fmt.Sprintf(messageRejected, reason)
	return result
}

// lexicalFailure monta o rastro de um só passo de uma sentença barrada pelo
// léxico: a coluna da entrada mostra os tokens lidos até ali e o caractere inválido.
func lexicalFailure(g *Grammar, start []string, input string, lexErr *LexicalError) Result {
	prefix := string([]rune(input)[:lexErr.Pos-1])
	read, _ := TokenizeFor(g, prefix)
	read = read[:len(read)-1] // tira o "$" do prefixo
	remaining := append(kinds(read), string(lexErr.Char), EndMarker)
	step := Step{
		Number: 1,
		Stack:  start,
		Input:  remaining,
		Action: ErrorPrefix + " léxico: " + lexErr.Detail(),
	}
	return Result{
		Message: fmt.Sprintf(messageRejected, lexErr.Error()),
		Tokens:  read,
		Steps:   []Step{step},
	}
}

// refuseConflicts monta o rastro de um só passo de uma sentença que não é
// analisada porque a gramática não é LL(1). A coluna da entrada mostra os
// tokens quando a sentença não tem erro léxico.
func refuseConflicts(g *Grammar, start []string, input string) Result {
	tokens, err := TokenizeFor(g, input)
	remaining := []string{EndMarker}
	if err == nil {
		remaining = kinds(tokens)
	}
	step := Step{Number: 1, Stack: start, Input: remaining, Action: ErrorPrefix + ": " + MessageNotLL1}
	return Result{Message: MessageNotLL1 + ".", Tokens: tokens, Steps: []Step{step}}
}

func kinds(tokens []Token) []string {
	result := make([]string, len(tokens))
	for i, token := range tokens {
		result[i] = token.Kind
	}
	return result
}

func describeSymbol(symbol string) string {
	if symbol == EndMarker {
		return "fim da entrada"
	}
	return "'" + symbol + "'"
}

// unexpected descreve o token que travou a análise: o lexema e a posição
// ajudam o usuário a achar o erro no que digitou.
func unexpected(token Token) string {
	if token.Kind == EndMarker {
		return "fim da entrada inesperado"
	}
	return fmt.Sprintf("token '%s' inesperado na posição %d", token.Lexeme, token.Pos)
}

// matchAction escreve "Casa 'id' (soma)", ou só "Casa '+'" quando o lexema é
// o próprio terminal.
func matchAction(token Token) string {
	if token.Lexeme == token.Kind {
		return fmt.Sprintf("Casa '%s'", token.Kind)
	}
	return fmt.Sprintf("Casa '%s' (%s)", token.Kind, token.Lexeme)
}
