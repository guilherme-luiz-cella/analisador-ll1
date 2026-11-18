---
class: web.app
path: web/app.js
kind: web
step: 2
---

# WebApp

The analyzer page (`web/index.html`, `app.js`, `style.css`) served at the site root. It loads `main.wasm`, lets the user pick a grammar in the "Gramática" dropdown (grouped with `<optgroup>`, remembered in the URL query as `?g=<id>`, so the section links in the hash do not drop it; an old `#g=<id>` link is rewritten once), shows the grammar, the FIRST and FOLLOW sets and the parsing table, takes a typed or generated sentence and shows the verdict with the tokens and the stack, step by step or all at once.

## Public functions
- Button handlers in `app.js`: Analisar, Gerar válida, Gerar inválida, Limpar, the example sentences of the selected grammar and the step controls (anterior, próximo, reproduzir).
- The dropdown: changing it re-renders grammar, FIRST/FOLLOW, LL(1) checks, table and examples and clears the analysis.

## Connections
- Calls [[Wasm]]: `globalThis.ll1.presets()`, `grammar(id)`, `sets(id)`, `table(id)`, `parse(input, id)`, `generate(id)` and `generateInvalid(id)` return JSON strings.
- Reads [[presetJSON]]: dropdown entries, "Fonte" line, description and example buttons.
- Reads [[resultJSON]]: tokens, steps and the verdict drawn on screen.
- Reads [[tableJSON]]: the parsing table; the cell a step consulted (from [[cellRefJSON]]) is highlighted.
- Reads [[grammarJSON]]: the productions and symbol lists shown in the grammar panel; the LL(1) condition panel uses the FIRST of each body from [[productionJSON]].
- Reads [[setsRowJSON]]: one row per non-terminal in the FIRST and FOLLOW table.

## Notes
All text from the parser is put on the page with `textContent`, never as HTML. The written explanation (why LL(1), requirements, derivation) belongs to the assignment grammar and is shown only when it is selected; other grammars show their description and source instead. A grammar with conflicts gets the verdict "NÃO ANALISADA" and its conflicting cells highlighted. The page links to [[GraphPage]] (`/graph`).
