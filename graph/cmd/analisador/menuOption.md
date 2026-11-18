---
class: analisador.menuOption
path: cmd/analisador/main.go
kind: cli
step: 1
---

# menuOption

One entry of the main menu: the key the user types, the label shown and the `app` method that runs. The action returns false when the input ended, so the loop can stop cleanly.

## Connections
- Dispatches to [[app]]: each entry is a method expression such as `(*app).showGrammar`.
