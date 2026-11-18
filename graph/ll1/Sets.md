---
class: ll1.symbolSet
path: ll1/sets.go
kind: support
step: 1
---

# Sets

FIRST and FOLLOW as a fixed-point computation. `symbolSet` (a `map[string]bool`) is the unordered set used while iterating; the public methods hand back slices in display order.

## Public functions
- `(*Grammar).First() map[string][]string`, `FirstOfSequence(seq) []string` and `Follow() map[string][]string`: see [[Grammar]]. FIRST is ordered like `Terminals` with ε last, FOLLOW like `Terminals` with `$` last.

## Connections
- Reads [[Grammar]]: productions and symbol order.
- Feeds [[Table]]: `BuildTable` uses the unexported `firstSets`, `firstOfSequence` and `followSets` to fill the cells.

## Notes
Each rule is repeated until no set grows, so the result never depends on the order of the productions.
