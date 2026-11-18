---
class: ll1.Table
path: ll1/table.go
kind: support
step: 1
---

# Table

The LL(1) parsing table `M[non-terminal, terminal]` holding production numbers. `Columns` are the terminals plus `$`; `Conflicts` is empty for an LL(1) grammar.

## Public functions
- `BuildTable(g *Grammar) *Table`: puts `A -> α` in `M[A, a]` for each `a` in FIRST(α) and, when ε is in FIRST(α), in `M[A, b]` for each `b` in FOLLOW(A). A cell claimed twice is recorded in `Conflicts` (Portuguese text) and in `ConflictCells`; the first production keeps it.
- `Production(nonTerminal, column) (Production, bool)`: the production in a cell; false for an error cell. Production numbers are 1-based positions in `Grammar.Productions`, so the lookup is by index.

## Connections
- Reads [[Grammar]]: the productions and symbol order.
- Uses [[Sets]]: FIRST and FOLLOW decide which cells are filled.
- Returns [[Production]]: from a cell lookup.
- Records [[ConflictCell]]: one per cell with more than one production.
