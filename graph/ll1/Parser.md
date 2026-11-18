---
class: ll1.Parse
path: ll1/parser.go
kind: support
step: 1
---

# Parser

The non-recursive predictive parser: an explicit stack driven by the parsing table.

## Public functions
- `Parse(t *Table, input string) Result`: refuses a table with conflicts (one error step, message `A gramática não é LL(1): a tabela tem conflitos.`); otherwise tokenizes with the lexer of the table's grammar, then loops: top `$` with input `$` accepts; a terminal on top must match the current token; a non-terminal is replaced by the body of `M[top, token]` pushed in reverse. Every action, including the final accept or error, is recorded as a [[Step]].

## Connections
- Uses [[Lexer]]: `TokenizeFor` splits the input with the grammar's terminals.
- Uses [[Table]]: picks the production for each non-terminal; its conflicts make the parser refuse.
- Reads [[LexicalError]]: turns it into a one-step rejected trace.
- Returns [[Result]]: verdict, message, tokens and steps.

## Notes
A lexical error yields a one-step trace (stack `[$ E]`, input = tokens read so far, the bad character and `$`) so the screen always shows a trace ending in an error.
