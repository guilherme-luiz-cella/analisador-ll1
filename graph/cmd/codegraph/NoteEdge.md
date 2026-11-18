---
class: codegraph.NoteEdge
path: cmd/codegraph/notes.go
kind: tool
step: 3
---

# NoteEdge

A `[[wikilink]]` from one note to another, as `from` and `to` ids.

## Connections
- Joins [[Note]]: both ends are note ids.

## Notes
Links inside code spans and fenced code are ignored, `[[Name|alias]]` and `[[Name#heading]]` point to `Name`, matching is case-insensitive and duplicates and self-links are dropped.
