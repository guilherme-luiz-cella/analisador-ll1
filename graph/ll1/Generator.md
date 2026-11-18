---
class: ll1.Generate
path: ll1/generator.go
kind: support
step: 1
---

# Generator

Random sentence generation for the terminal menu options 5 and 6 and for the web buttons, for any grammar.

## Public functions
- `Generate(g, rng, maxDepth) string`: random leftmost derivation from the start symbol; from `maxDepth` on, only the productions with the shortest derivation height are chosen, so it always ends, even for the left-recursive `E -> E + T | T`. Lexemes: `x y z a b soma total` and numbers 0..99, joined by spaces. Assumes every non-terminal derives some sentence (true for every preset).
- `GenerateInvalid(g, rng) string`: mutates a valid sentence (drop a token, duplicate an operator, add an unbalanced parenthesis or insert an illegal character) and retries until [[Parser]] rejects it. An empty result (dropping the only token of `x`) is skipped, so the sentence is never blank. "Operators" are the one-character terminals of `g`; illegal characters that are terminals of `g` (such as `;` or `#`) are skipped.

## Connections
- Uses [[Grammar]]: the productions to derive from.
- Calls [[Parser]]: to confirm a mutated sentence is really rejected.
- Builds [[Table]]: `GenerateInvalid` needs one to run the parser.

## Notes
`DefaultMaxDepth = 4` keeps sentences short enough to read the trace on a projector; the terminal, the web bridge and `GenerateInvalid` all use it.
