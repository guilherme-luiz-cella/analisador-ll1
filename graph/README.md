# Mapa de notas

Um mapa do projeto em Markdown: uma nota por tipo ou ponto de entrada do código, ligadas por `[[wikilinks]]`. A pasta pode ser aberta no Obsidian (visão de grafo) e é lida também pela página `/graph` do site, que a desenha no mesmo estilo.

O código vale mais que a nota: se os dois divergirem, corrija a nota.

## O que ganha uma nota

Todo tipo (`struct`, `interface`) e todo ponto de entrada de `ll1/`, `cmd/analisador`, `cmd/wasm` e `cmd/codegraph`, mais uma nota para a página web (`web/app.js`) e uma para a página do grafo. Testes e arquivos gerados não têm nota.

## Local e nome

O caminho espelha o do código: `ll1/parser.go` fica descrito em `graph/ll1/Parser.md`. O nome do arquivo é o nome da nota e o que vai dentro de `[[ ]]`; dois arquivos não podem ter o mesmo nome, mesmo em pastas diferentes.

## Formato

```markdown
---
class: ll1.Parse
path: ll1/parser.go
kind: support
step: 1
---

# Parser

O que faz e por que existe, em uma a três frases.

## Public functions
- `Parse(t *Table, input string) Result`: o que faz e quando falha.

## Connections
- Uses [[Table]]: escolhe a produção de cada não terminal.
- Returns [[Result]]: veredito, mensagem e passos.

## Notes
O que o código não mostra: uma regra, uma armadilha. Omitir se vazio.
```

- `class`: nome qualificado (`pacote.Nome`). `path`: arquivo de origem, relativo à raiz do módulo; a ferramenta confere que ele existe.
- `kind`: `support` (pacote `ll1`), `cli` (terminal), `wasm` (ponte para o navegador), `tool` (ferramentas de linha de comando) ou `web` (páginas). Um tipo novo precisa ser acrescentado aqui e em `cmd/codegraph/notes.go`.
- `step`: número da etapa do projeto em que a nota nasceu.

## Regras

- Só ligações de saída, e cada uma diz por que existe. Para ver quem usa uma nota, use os backlinks do Obsidian ou a busca por `[[Nome]]`; uma lista de "usado por" feita à mão fica velha.
- Só ligue para notas que existem. Pacotes externos e a biblioteca padrão vão em `código`.
- Liste só funções públicas; não descreva funções privadas nem repita o código linha por linha.

## Gerar o JSON da página

```
go run ./cmd/codegraph -notes graph -o web/graph/notes.json
```

O comando lê todas as notas (menos este README), monta o grafo e falha se houver ligação para nota inexistente, nome repetido, `kind` desconhecido ou `path` que não existe. O `build-site.ps1` faz isso ao montar o site.
