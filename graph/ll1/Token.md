---
class: ll1.Token
path: ll1/lexer.go
kind: support
step: 1
---

# Token

One lexical unit: terminal kind (`id`, `num`, `+`, ... or `$`), the lexeme as typed and its 1-based column.

## Notes
Columns count runes, not bytes, so accented input still reports the right position. Tokens are created by [[Lexer]].
