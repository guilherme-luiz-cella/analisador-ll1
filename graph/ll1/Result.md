---
class: ll1.Result
path: ll1/parser.go
kind: support
step: 1
---

# Result

Outcome of [[Parser]]: accepted flag, final Portuguese message (`Sentença ACEITA...` or `Sentença REJEITADA: motivo.`), the tokens and the trace.

## Connections
- Holds [[Token]]: the tokens read (only those before the bad character on a lexical error).
- Holds [[Step]]: the full trace, first stack `[$ E]`, last action `Aceita` or `Erro...`.
