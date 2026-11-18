---
class: ll1.Grammar
path: ll1/grammar.go
kind: support
step: 1
---

# Grammar

The fixed expression grammar: start symbol, non-terminals, terminals and the productions in display order. FIRST and FOLLOW live in `ll1/sets.go`; the constants `Epsilon` (`ε`), `EndMarker` (`$`), `TerminalID` and `TerminalNum` are defined here.

## Public functions
- `Default() *Grammar`: the assignment grammar, productions 1..11.
- `IsTerminal(s) bool`, `IsNonTerminal(s) bool`: symbol membership.
- `ProductionsOf(nonTerminal) []Production`: alternatives of one non-terminal, in order.
- `String() string`: one line per non-terminal, alternatives joined by ` | `.
- `First()`, `FirstOfSequence(seq)`, `Follow()`: the sets, implemented in [[Sets]].

## Connections
- Holds [[Production]]: the rules it is made of.
- Computes through [[Sets]]: FIRST and FOLLOW.

## Notes
FIRST and FOLLOW are computed by the fixed-point algorithm, never hard-coded; the tests compare them against the textbook table.
