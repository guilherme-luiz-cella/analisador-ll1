---
class: analisador.app
path: cmd/analisador/main.go
kind: cli
step: 1
---

# app

The terminal program: pt-BR menu (1 Gramática, 2 FIRST e FOLLOW, 3 Tabela, 4 Digitar, 5 Gerar válida, 6 Gerar inválida, 7 Trocar gramática, 0 Sair). The menu header names the current grammar; option 7 lists the presets by group, with author, source and URL, and switches by number (Enter keeps the current one). Options 4 to 6 print the tokens, the step table (Passo, Pilha, Entrada, Ação) and the result, then wait for Enter and show the menu again.

## Public functions
- `run(in, out, colors, rng)`: the menu loop; ends on 0 or at the end of the input. Used by `main` and by the tests.

## Connections
- Reads [[menuOption]]: the table of menu entries it dispatches on.
- Draws with [[palette]]: boxes, aligned tables and optional ANSI colors.
- Calls [[Console]]: `prepareConsole` decides whether colors can be used.
- Uses [[Preset]]: starts with the assignment grammar; option 7 switches.
- Uses [[Grammar]]: grammar, FIRST and FOLLOW sets for options 1 and 2.
- Uses [[Table]]: option 3 prints it.
- Calls [[Parser]]: options 4 to 6 trace the recognition.
- Calls [[Generator]]: options 5 and 6 create the sentence.

## Notes
Frames and tables are drawn by `render.go`. Colors are ANSI, used only when `NO_COLOR` is unset and the console accepts them; piped output is plain.
