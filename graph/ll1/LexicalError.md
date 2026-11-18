---
class: ll1.LexicalError
path: ll1/lexer.go
kind: support
step: 1
---

# LexicalError

Error returned by `Tokenize` for a character that starts no token. It carries the character and its 1-based column.

## Public functions
- `Error() string`: `erro léxico: caractere '#' inválido na posição 3`.
- `Detail() string`: the same text without the `erro léxico:` prefix, used in the trace action.
