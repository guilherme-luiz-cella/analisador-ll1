---
class: ll1.Production
path: ll1/grammar.go
kind: support
step: 1
---

# Production

One numbered grammar rule (`Head -> Body`). Numbers are 1-based and are what the cells of the parsing table hold.

## Public functions
- `String() string`: prints `E' -> + T E'`.
- `IsEmpty() bool`: true for the ε production (`Body` is `[ε]`).

## Notes
The empty production keeps `Body = []string{Epsilon}` so it prints as `T' -> ε`; the parser pushes nothing for it.
