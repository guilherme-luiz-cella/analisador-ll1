---
class: ll1.ConflictCell
path: ll1/table.go
kind: support
step: 6
---

# ConflictCell

A cell of the [[Table]] claimed by more than one production: the non-terminal, the column and the numbers of every production that claimed it, in order (the first one keeps the cell). The web uses it to highlight the cell and print all the productions in it.

## Connections
- Recorded by [[Table]]: `BuildTable` adds one per conflicting cell, next to the text in `Conflicts`.
