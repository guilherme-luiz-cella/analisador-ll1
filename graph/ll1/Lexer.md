---
class: ll1.Tokenize
path: ll1/lexer.go
kind: support
step: 1
---

# Lexer

The scanner of the sentence, in `ll1/lexer.go`: turns the typed text into [[Token]] values before the parser sees it.

## Public functions
- `Tokenize(input string) ([]Token, error)`: identifiers `[A-Za-z_][A-Za-z0-9_]*` become `id`, digit runs `num`, `+ - * / ( )` stand for themselves and spaces, tabs and line breaks are skipped. The last token is always `$`. Any other character stops the scan with a [[LexicalError]].

## Connections
- Produces [[Token]]: one per lexeme, plus the `$` end marker.
- Returns [[LexicalError]]: for the first character that starts no token.

## Notes
The input is read as runes, so the reported column is right for accented text.
