---
class: ll1.Parse
path: ll1/parser.go
kind: support
step: 1
---

# Parser

The non-recursive predictive parser: an explicit stack driven by the parsing table.

## Public functions
- `Parse(t *Table, input string) Result`: tokenizes, then loops: top `$` with input `$` accepts; a terminal on top must match the current token; a non-terminal is replaced by the body of `M[top, token]` pushed in reverse. Every action, including the final accept or error, is recorded as a [[Step]].

## Connections
- Uses [[Lexer]]: `Tokenize` splits the input.
- Uses [[Table]]: picks the production for each non-terminal.
- Reads [[LexicalError]]: turns it into a one-step rejected trace.
- Returns [[Result]]: verdict, message, tokens and steps.

## Notes
A lexical error yields a one-step trace (stack `[$ E]`, input = tokens read so far, the bad character and `$`) so the screen always shows a trace ending in an error.
