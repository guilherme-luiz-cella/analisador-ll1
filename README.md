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

A linguagem reconhecida é a das expressões aritméticas com identificadores, números inteiros, os operadores `+ - * /` e parênteses.

### Definição formal

G = (N, T, P, S), com:

- N = { E, E', T, T', F } (não terminais)
- T = { id, num, +, -, *, /, (, ) } (terminais; `$` marca o fim da entrada)
- P = as 11 produções numeradas abaixo
- S = E (símbolo inicial)

### Produções

Os números são os que aparecem nas células da tabela de parsing.

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

### Por que é LL(1)

**Fatorada.** Nenhum par de alternativas do mesmo não terminal começa pelo mesmo símbolo: `E'` escolhe por `+` ou `-`, `T'` por `*` ou `/`, `F` por `(`, `id` ou `num`. Olhar um único token basta para decidir.

**Sem recursão à esquerda.** A gramática usual de expressões é recursiva à esquerda:

```
E -> E + T | E - T | T
T -> T * F | T / F | F
F -> ( E ) | id | num
```

Um analisador descendente que expande `E -> E + T` volta a ter `E` no topo da pilha sem consumir nada e entra em laço infinito. A regra `A -> A α | β` é trocada por `A -> β A'` e `A' -> α A' | ε`. Aplicada a `E` (com `α = + T` ou `- T`, `β = T`) e a `T` (com `α = * F` ou `/ F`, `β = F`), ela dá:

```
E  -> T E'
E' -> + T E' | - T E' | ε
T  -> F T'
T' -> * F T' | / F T' | ε
```

**Não ambígua.** Toda gramática LL(1) é não ambígua, e a tabela de parsing não tem nenhuma célula com duas produções. A precedência está nos níveis: `+` e `-` ficam em `E'`, acima de `*` e `/`, que ficam em `T'`. Em `a + b * c`, a multiplicação fica mais funda na árvore e é feita primeiro. A associatividade continua à esquerda: `E'` e `T'` geram os operadores da esquerda para a direita, e `a - b - c` vale `(a - b) - c`, como na gramática original.

### Condição LL(1) em cada regra

Para cada `A -> α | β`: FIRST(α) ∩ FIRST(β) = ∅. Quando uma alternativa deriva ε: FIRST(outras alternativas) ∩ FOLLOW(A) = ∅. Os conjuntos usados são os da seção seguinte.

| Regra | Verificação | Resultado |
|---|---|---|
| E | uma única produção (1): não há escolha | satisfeita |
| E' | FIRST(+ T E') ∩ FIRST(- T E') = { + } ∩ { - } | ∅ |
| E' | (FIRST(+ T E') ∪ FIRST(- T E')) ∩ FOLLOW(E') = { +, - } ∩ { ), $ } | ∅ |
| T | uma única produção (5): não há escolha | satisfeita |
| T' | FIRST(* F T') ∩ FIRST(/ F T') = { * } ∩ { / } | ∅ |
| T' | (FIRST(* F T') ∪ FIRST(/ F T')) ∩ FOLLOW(T') = { *, / } ∩ { +, -, ), $ } | ∅ |
| F | FIRST(( E )) ∩ FIRST(id) = { ( } ∩ { id } | ∅ |
| F | FIRST(( E )) ∩ FIRST(num) = { ( } ∩ { num } | ∅ |
| F | FIRST(id) ∩ FIRST(num) = { id } ∩ { num } | ∅ |

A versão web calcula esta mesma verificação a partir dos conjuntos que o programa gera.

### Requisitos mínimos do trabalho

- Pelo menos 4 regras: são 5 não terminais (E, E', T, T', F) e 11 produções.
- Regras com 2 ou mais produções: E', T' e F, com 3 cada.
- Regras com produção vazia: E' (produção 4) e T' (produção 8).
- Fatorada, sem recursão à esquerda e não ambígua, como mostrado acima.

### Derivação mais à esquerda de `id + num * id`

O número à direita é a produção aplicada ao não terminal mais à esquerda. É a mesma sequência que o analisador aplica ao reconhecer `x + 1 * y`.

```
E
=> T E'                 (1)
=> F T' E'              (5)
=> id T' E'             (10)
=> id E'                (8)
=> id + T E'            (2)
=> id + F T' E'         (5)
=> id + num T' E'       (11)
=> id + num * F T' E'   (6)
=> id + num * id T' E'  (10)
=> id + num * id E'     (8)
=> id + num * id        (4)
```

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

## Outras gramáticas

A gramática do trabalho é a das expressões, e ela é a seleção padrão. Para demonstração, o programa traz mais 19 gramáticas fixas, escolhidas no menu "Gramática" do site (o endereço guarda a escolha, por exemplo `#g=json`) ou na opção 7 do terminal. O usuário não digita gramáticas.

- Exemplos próprios: expressões booleanas, declarações de variáveis, listas aninhadas e um exemplo NÃO LL(1) (`E -> E + T | T`). Neste último a tabela mostra o conflito em `M[E, id]` e o analisador se recusa a analisar.
- Da literatura e da web, com autor e fonte na tela: JSON (RFC 4627), S-expressões (McCarthy), PL/0 (Wirth), programas em linha reta (Appel), TINY (Louden), expressões regulares (Might), Dyck, XML (W3C), CSV (RFC 4180), lógica proposicional (IEP), cálculo lambda (SEP), tabelas Lua (manual do Lua 5.4), notação de Wirth (EBNF), Brainf*ck e declarações Pascal (ISO 7185). Quase todas foram adaptadas para LL(1) (fatoração, remoção de recursão à esquerda, repetições viradas listas com ε); o processo e os conjuntos calculados à mão estão em `steps/06-presets-research.md`.

O analisador léxico segue a gramática escolhida: palavras iguais a um terminal (`and`, `var`, `if`) viram esse terminal, outras palavras viram `id`, dígitos viram `num`, e cada terminal de um caractere vale por si. Não há tokens de dois caracteres (`:=` virou `=`, `..` virou `. .`), e `|` não é terminal porque separa alternativas na notação.

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
7  Trocar gramática
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
