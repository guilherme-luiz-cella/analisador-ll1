package ll1

import "fmt"

// Step is one row of the parsing trace.
type Step struct {
	Number int
	Stack  []string // bottom first, so the top is the last element; starts as ["$", "E"]
	Input  []string // remaining token kinds, current first, ends with "$"
	Action string   // Portuguese, e.g. "E -> T E'", "Casa 'id' (soma)", "Aceita", "Erro: ..."
}

// Result is the outcome of parsing one sentence.
type Result struct {
	Accepted bool
	Message  string // Portuguese final message
	Tokens   []Token
	Steps    []Step
}

// Fixed texts of the trace and of the final message.
const (
	ActionAccept    = "Aceita"
	ErrorPrefix     = "Erro"
	MessageAccepted = "Sentença ACEITA: a entrada foi reconhecida pela gramática."
	messageRejected = "Sentença REJEITADA: %s."
)

// Parse runs the non-recursive predictive parser over input, recording a
// Step for every action, including the final accept or error.
func Parse(t *Table, input string) Result {
	start := []string{EndMarker, t.Grammar.Start}
	tokens, err := Tokenize(input)
	if err != nil {
		// Tokenize only fails with a *LexicalError.
		return lexicalFailure(start, input, err.(*LexicalError))
	}

	result := Result{Tokens: tokens}
	stack := start
	position := 0
	for {
		top := stack[len(stack)-1]
		current := tokens[position]
		step := Step{
			Number: len(result.Steps) + 1,
			Stack:  append([]string{}, stack...),
			Input:  kinds(tokens[position:]),
		}

		switch {
		case top == EndMarker && current.Kind == EndMarker:
			step.Action = ActionAccept
			result.Steps = append(result.Steps, step)
			result.Accepted = true
			result.Message = MessageAccepted
			return result

		case !t.Grammar.IsNonTerminal(top):
			if top != current.Kind {
				reason := fmt.Sprintf("%s, esperado %s", unexpected(current), describeSymbol(top))
				return reject(result, step, reason)
			}
			step.Action = matchAction(current)
			stack = stack[:len(stack)-1]
			position++

		default:
			p, ok := t.Production(top, current.Kind)
			if !ok {
				reason := fmt.Sprintf("%s: M[%s, %s] vazia", unexpected(current), top, current.Kind)
				return reject(result, step, reason)
			}
			step.Action = p.String()
			stack = stack[:len(stack)-1]
			if !p.IsEmpty() {
				for i := len(p.Body) - 1; i >= 0; i-- {
					stack = append(stack, p.Body[i])
				}
			}
		}
		result.Steps = append(result.Steps, step)
	}
}

func reject(result Result, step Step, reason string) Result {
	step.Action = ErrorPrefix + ": " + reason
	result.Steps = append(result.Steps, step)
	result.Message = fmt.Sprintf(messageRejected, reason)
	return result
}

// lexicalFailure builds the single-step trace of a sentence rejected by the
// lexer: the input column shows the tokens read so far and the bad character.
func lexicalFailure(start []string, input string, lexErr *LexicalError) Result {
	prefix := string([]rune(input)[:lexErr.Pos-1])
	read, _ := Tokenize(prefix)
	read = read[:len(read)-1] // drop the "$" of the prefix
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

func unexpected(token Token) string {
	if token.Kind == EndMarker {
		return "fim da entrada inesperado"
	}
	return fmt.Sprintf("token '%s' inesperado na posição %d", token.Lexeme, token.Pos)
}

// matchAction reads "Casa 'id' (soma)", or just "Casa '+'" when the lexeme
// is the terminal itself.
func matchAction(token Token) string {
	if token.Lexeme == token.Kind {
		return fmt.Sprintf("Casa '%s'", token.Kind)
	}
	return fmt.Sprintf("Casa '%s' (%s)", token.Kind, token.Lexeme)
}
