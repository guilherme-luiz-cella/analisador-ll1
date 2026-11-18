# Analisador Sintático LL(1)

Analisador sintático descendente preditivo tabular (não recursivo) LL(1), feito em Go para o TDE de Compiladores (URI Erechim, Prof. Fabio Zanin).

- Versão web: https://com.cella.website
- Grafo do código: https://com.cella.website/graph

## O que o programa faz

- Mostra a gramática, os conjuntos FIRST e FOLLOW e a tabela de parsing.
- Aceita uma sentença digitada ou gera sentenças automaticamente (válidas e inválidas).
- Mostra cada passo do reconhecimento: pilha, entrada restante e ação.
- Rejeita sentenças com erro sintático ou léxico, indicando o motivo e a posição.
- Ao final mostra o resultado (ACEITA ou REJEITADA) e volta ao menu para recomeçar.

## Gramática

```
E  -> T E'
E' -> + T E' | - T E' | ε
T  -> F T'
T' -> * F T' | / F T' | ε
F  -> ( E ) | id | num
```

Símbolo inicial `E`. Terminais: `id num + - * / ( )`, fim da entrada `$`. A gramática está fatorada, sem recursão à esquerda e não é ambígua.

Produções numeradas (os números aparecem na tabela):

| Nº | Produção |
|---|---|
| 1 | E -> T E' |
| 2 | E' -> + T E' |
| 3 | E' -> - T E' |
| 4 | E' -> ε |
| 5 | T -> F T' |
| 6 | T' -> * F T' |
| 7 | T' -> / F T' |
| 8 | T' -> ε |
| 9 | F -> ( E ) |
| 10 | F -> id |
| 11 | F -> num |

## Conjuntos FIRST e FOLLOW

Calculados pelo programa (algoritmo de ponto fixo):

| Não terminal | FIRST | FOLLOW |
|---|---|---|
| E | { id, num, ( } | { ), $ } |
| E' | { +, -, ε } | { ), $ } |
| T | { id, num, ( } | { +, -, ), $ } |
| T' | { *, /, ε } | { +, -, ), $ } |
| F | { id, num, ( } | { +, -, *, /, ), $ } |

## Tabela de parsing

Células vazias indicam erro. Não há conflitos, portanto a gramática é LL(1).

| M | id | num | + | - | * | / | ( | ) | $ |
|---|---|---|---|---|---|---|---|---|---|
| E | 1 | 1 | | | | | 1 | | |
| E' | | | 2 | 3 | | | | 4 | 4 |
| T | 5 | 5 | | | | | 5 | | |
| T' | | | 8 | 8 | 6 | 7 | | 8 | 8 |
| F | 10 | 11 | | | | | 9 | | |

## Analisador léxico

- Identificadores `[A-Za-z_][A-Za-z0-9_]*` viram `id`.
- Sequências de dígitos viram `num`.
- `+ - * / ( )` são os próprios terminais; espaços são ignorados.
- Qualquer outro caractere é erro léxico, e a sentença é rejeitada com a posição do caractere.

## Como executar

No Windows, dê dois cliques em `executar.bat` (ele ativa UTF-8 no console para mostrar o ε e chama `analisador.exe`).

Com o Go instalado:

```
go run ./cmd/analisador
```

Menu:

```
1  Gramática
2  Conjuntos FIRST e FOLLOW
3  Tabela de parsing
4  Digitar sentença
5  Gerar sentença válida
6  Gerar sentença inválida
0  Sair
```

Exemplos para digitar: `a+b*c`, `(a+b)*c`, `a-b/c+d*(e-1)` (aceitas); `a+`, `(a`, `a**b`, `a+#` (rejeitadas).

Para desativar as cores, defina a variável de ambiente `NO_COLOR`.

## Desenvolvimento

```
go test ./...
go vet ./...
go build -o dist/analisador.exe ./cmd/analisador
```

Estrutura:

- `ll1/`: gramática, FIRST e FOLLOW, tabela, analisador léxico, analisador sintático com pilha e gerador de sentenças.
- `cmd/analisador/`: programa de terminal.
