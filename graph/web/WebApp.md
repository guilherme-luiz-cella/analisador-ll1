---
class: web.app
path: web/app.js
kind: web
step: 2
---

# WebApp

The analyzer page (`web/index.html`, `app.js`, `style.css`) served at the site root. It loads `main.wasm`, shows the grammar, the FIRST and FOLLOW sets and the parsing table, takes a typed or generated sentence and shows the verdict with the tokens and the stack, step by step or all at once.

## Public functions
- Button handlers in `app.js`: Analisar, Gerar válida, Gerar inválida, Limpar, the example sentences and the step controls (anterior, próximo, reproduzir).

## Connections
- Calls [[Wasm]]: `globalThis.ll1.grammar()`, `sets()`, `table()`, `parse()`, `generate()` and `generateInvalid()` return JSON strings.
- Reads [[resultJSON]]: tokens, steps and the verdict drawn on screen.
- Reads [[tableJSON]]: the parsing table; the cell a step consulted (from [[cellRefJSON]]) is highlighted.
- Reads [[grammarJSON]]: the productions and symbol lists shown in the grammar panel.
- Reads [[setsRowJSON]]: one row per non-terminal in the FIRST and FOLLOW table.

## Notes
All text from the parser is put on the page with `textContent`, never as HTML. The page links to [[GraphPage]] (`/graph`).
