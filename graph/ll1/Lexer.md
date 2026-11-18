---
class: ll1.Tokenize
path: ll1/lexer.go
kind: support
step: 1
---

# Lexer

The scanner of the sentence, in `ll1/lexer.go`: turns the typed text into [[Token]] values before the parser sees it. It follows the grammar being parsed.

## Public functions
- `TokenizeFor(g *Grammar, input string) ([]Token, error)`: a word `[A-Za-z_][A-Za-z0-9_]*` equal to a terminal of `g` (a keyword such as `and`, `var`, `int`) becomes that terminal and any other word becomes `id`; digit runs become `num`; a character that is a one-character terminal of `g` stands for itself; spaces, tabs and line breaks are skipped. The last token is always `$`. Any other character stops the scan with a [[LexicalError]].
- `Tokenize(input string) ([]Token, error)`: `TokenizeFor` with the assignment grammar, so `+ - * / ( )` are the only symbols.

## Connections
- Reads [[Grammar]]: its terminals decide keywords and symbols.
- Produces [[Token]]: one per lexeme, plus the `$` end marker.
- Returns [[LexicalError]]: for the first character that starts no token.

## Notes
The input is read as runes, so the reported column is right for accented text. The words `id` and `num` are never keywords: typed, they are identifiers, as in the assignment lexer. There are no two-character tokens.
