package ll1

import "strings"

// PresetGroup diz de onde vem uma gramática pronta; a web agrupa a lista de
// escolha por ele e o terminal agrupa a sua lista.
type PresetGroup int

const (
	GroupAssignment PresetGroup = iota // a gramática avaliada no trabalho
	GroupExamples                      // exemplos pequenos escritos para este projeto
	GroupLiterature                    // gramáticas publicadas, adaptadas para LL(1)
)

// Label é o título do grupo mostrado na tela.
func (g PresetGroup) Label() string {
	switch g {
	case GroupAssignment:
		return "Trabalho"
	case GroupExamples:
		return "Exemplos"
	default:
		return "Da literatura e da web"
	}
}

// DefaultPresetID é o id da gramática do trabalho, a escolhida por padrão.
const DefaultPresetID = "expressoes"

// Preset é uma gramática pronta que o usuário pode escolher. Examples estão
// na linguagem (são aceitos, menos numa gramática com conflitos, que recusa
// tudo); InvalidExamples têm de ser rejeitados.
type Preset struct {
	ID              string
	Name            string
	Description     string
	Group           PresetGroup
	Author          string // vazio nas gramáticas do próprio projeto
	Source          string // título da publicação
	URL             string
	Adapted         bool // a gramática publicada foi alterada para ser LL(1)
	Grammar         *Grammar
	Examples        []string
	InvalidExamples []string
}

// Presets lista todas as gramáticas prontas, a do trabalho primeiro. Cada
// chamada monta gramáticas novas, então quem chama pode guardá-las ou mudá-las.
func Presets() []Preset {
	return []Preset{
		{
			ID:              DefaultPresetID,
			Name:            "Expressões aritméticas (gramática do trabalho)",
			Description:     "A gramática do trabalho: expressões com identificadores, números, + - * / e parênteses, já fatorada e sem recursão à esquerda.",
			Group:           GroupAssignment,
			Grammar:         Default(),
			Examples:        []string{"a+b*c", "(x+1)*y"},
			InvalidExamples: []string{"a+", "2 $ 3"},
		},
		{
			ID:          "booleanas",
			Name:        "Expressões booleanas",
			Description: "Expressões com or, and e not, constantes true e false, identificadores e parênteses. Mesma ideia da gramática do trabalho: and tem precedência sobre or.",
			Group:       GroupExamples,
			Grammar: grammarFrom(`
				B  -> T B'
				B' -> or T B' | ε
				T  -> F T'
				T' -> and F T' | ε
				F  -> not F | ( B ) | true | false | id`),
			Examples:        []string{"a and not b or true", "not ( x or y ) and false", "true"},
			InvalidExamples: []string{"a and or b", "( a or b"},
		},
		{
			ID:          "declaracoes",
			Name:        "Declarações de variáveis",
			Description: "Uma declaração no estilo C: um tipo (int, float ou char), uma lista de nomes separados por vírgula e o ponto e vírgula no fim.",
			Group:       GroupExamples,
			Grammar: grammarFrom(`
				D  -> T L ;
				T  -> int | float | char
				L  -> id L'
				L' -> , id L' | ε`),
			Examples:        []string{"int a, b, c;", "float total;"},
			InvalidExamples: []string{"int a b;", "float a, b"},
		},
		{
			ID:          "listas",
			Name:        "Listas aninhadas",
			Description: "Listas entre colchetes, com números, identificadores ou outras listas separados por vírgula. A lista vazia é permitida.",
			Group:       GroupExamples,
			Grammar: grammarFrom(`
				L  -> [ E ]
				E  -> X E' | ε
				E' -> , X E' | ε
				X  -> num | id | L`),
			Examples:        []string{"[ ]", "[1, x, [2, [ ]]]"},
			InvalidExamples: []string{"[1, ]", "[1 2]"},
		},
		{
			ID:   "nao-ll1",
			Name: "Exemplo NÃO LL(1) (recursão à esquerda)",
			Description: "Gramática recursiva à esquerda, para mostrar a detecção de conflitos. As duas produções de E começam por id " +
				"(FIRST(E + T) = FIRST(T) = { id }), então a célula M[E, id] recebe as produções 1 e 2 e o analisador não tem como escolher " +
				"olhando um só token. Por isso ele se recusa a analisar. A correção é a mesma da gramática do trabalho: E → T E' e E' → + T E' | ε.",
			Group: GroupExamples,
			Grammar: grammarFrom(`
				E -> E + T | T
				T -> id`),
			Examples:        []string{"a + b + c", "x"},
			InvalidExamples: []string{"a +"},
		},
		{
			ID:          "json",
			Name:        "JSON",
			Description: "Valores JSON: objetos, arrays, números, true, false e null. Strings viram identificadores, e as partes opcionais e repetidas da RFC viraram listas com ε.",
			Group:       GroupLiterature,
			Author:      "Douglas Crockford",
			Source:      "RFC 4627: The application/json Media Type for JavaScript Object Notation (JSON), 2006",
			URL:         "https://www.rfc-editor.org/rfc/rfc4627",
			Adapted:     true,
			Grammar: grammarFrom(`
				Value     -> Object | Array | id | num | true | false | null
				Object    -> { Members }
				Members   -> Member Members' | ε
				Members'  -> , Member Members' | ε
				Member    -> id : Value
				Array     -> [ Elements ]
				Elements  -> Value Elements' | ε
				Elements' -> , Value Elements' | ε`),
			Examples:        []string{"{ }", "{ a : 1 , b : [ true , null ] }", "[ 1 , { x : false } , [ ] ]"},
			InvalidExamples: []string{"{ a : 1 , }", "[ 1 2 ]"},
		},
		{
			ID:          "s-expressoes",
			Name:        "Expressões simbólicas (Lisp)",
			Description: "As S-expressões de McCarthy: átomos, pares pontuados (a . b) e listas (a , b , c). As três formas começam com ( S e foram fatoradas à esquerda; o ponto · virou .",
			Group:       GroupLiterature,
			Author:      "John McCarthy",
			Source:      "Recursive Functions of Symbolic Expressions and Their Computation by Machine, Part I, Communications of the ACM, 1960",
			URL:         "http://www-formal.stanford.edu/jmc/recursive.pdf",
			Adapted:     true,
			Grammar: grammarFrom(`
				S    -> id | ( S Rest
				Rest -> ) | . S ) | , S Rest`),
			Examples:        []string{"a", "( a . b )", "( a , ( b . c ) , d )"},
			InvalidExamples: []string{"( )", "( a b )"},
		},
		{
			ID:          "pl0",
			Name:        "PL/0 (subconjunto)",
			Description: "A linguagem didática de Wirth: declaração var, atribuição, call, begin/end, if e while, terminando com ponto. := virou = e as repetições { } viraram listas com ε.",
			Group:       GroupLiterature,
			Author:      "Niklaus Wirth",
			Source:      "Algorithms + Data Structures = Programs, 1976 (gramática PL/0 reproduzida na Wikipédia)",
			URL:         "https://en.wikipedia.org/wiki/PL/0",
			Adapted:     true,
			Grammar: grammarFrom(`
				Program -> Var Stmt .
				Var     -> var id Ids ; | ε
				Ids     -> , id Ids | ε
				Stmt    -> id = Expr | call id | begin Stmt Stmts end | if Cond then Stmt | while Cond do Stmt | ε
				Stmts   -> ; Stmt Stmts | ε
				Cond    -> odd Expr | Expr Rel Expr
				Rel     -> = | # | < | >
				Expr    -> Term Expr'
				Expr'   -> + Term Expr' | - Term Expr' | ε
				Term    -> id | num | ( Expr )`),
			Examples:        []string{"var x , y ; begin x = 1 ; y = x + 2 end .", "while odd x do x = x - 1 .", "if x # 0 then call p ."},
			InvalidExamples: []string{"var x ; x = .", "begin x = 1 ; ."},
		},
		{
			ID:          "linha-reta",
			Name:        "Programas em linha reta (Appel)",
			Description: "A linguagem do capítulo 1 de Appel: atribuições, print e expressões com sequência ( comandos , expressão ). A original é ambígua e recursiva à esquerda; aqui os operadores são uma lista sem precedência.",
			Group:       GroupLiterature,
			Author:      "Andrew W. Appel",
			Source:      "Modern Compiler Implementation in C, Cambridge University Press, 1998, Grammar 1.3",
			URL:         "https://beckassets.blob.core.windows.net/product/readingsample/637255/9780521607650_excerpt_001.pdf",
			Adapted:     true,
			Grammar: grammarFrom(`
				Stm      -> Simple Stm'
				Stm'     -> ; Simple Stm' | ε
				Simple   -> id = Exp | print ( ExpList )
				Exp      -> Atom Exp'
				Exp'     -> Binop Atom Exp' | ε
				Atom     -> id | num | ( Stm , Exp )
				ExpList  -> Exp ExpList'
				ExpList' -> , Exp ExpList' | ε
				Binop    -> + | - | * | /`),
			Examples:        []string{"a = 5 + 3 ; b = ( print ( a , a - 1 ) , 10 * a ) ; print ( b )", "x = 1", "print ( 1 , 2 , x / 2 )"},
			InvalidExamples: []string{"a = 5 + ; print ( a )", "print ( )"},
		},
		{
			ID:          "tiny",
			Name:        "Linguagem TINY (Louden)",
			Description: "A linguagem TINY do livro de Louden: if/then/else/end, repeat/until, read, write e atribuição. Saíram * e /; := virou =.",
			Group:       GroupLiterature,
			Author:      "Kenneth C. Louden",
			Source:      "Compiler Construction: Principles and Practice, PWS, 1997 (parse.c do livro)",
			URL:         "https://faculty.chas.uni.edu/~wallingf/teaching/cs4550/support/tiny/parse.c",
			Adapted:     true,
			Grammar: grammarFrom(`
				StmtSeq  -> Stmt StmtSeq'
				StmtSeq' -> ; Stmt StmtSeq' | ε
				Stmt     -> if Exp then StmtSeq Else end | repeat StmtSeq until Exp | id = Exp | read id | write Exp
				Else     -> else StmtSeq | ε
				Exp      -> Simple Exp'
				Exp'     -> < Simple | = Simple | ε
				Simple   -> Factor Simple'
				Simple'  -> + Factor Simple' | - Factor Simple' | ε
				Factor   -> num | id | ( Exp )`),
			Examples: []string{
				"read x ; if 0 < x then f = 1 ; repeat f = f + x ; x = x - 1 until x = 0 ; write f end",
				"write 1",
				"if x then read y else write ( y + 1 ) end",
			},
			InvalidExamples: []string{"read x ;", "if x then write x"},
		},
		{
			ID:          "regex",
			Name:        "Expressões regulares",
			Description: "Concatenação, estrela * e alternância, com caracteres como identificadores. A alternância | virou +, como em Hopcroft e Ullman, porque | separa alternativas na notação da gramática.",
			Group:       GroupLiterature,
			Author:      "Matt Might",
			Source:      "Parsing regular expressions with recursive descent",
			URL:         "https://matt.might.net/articles/parsing-regex-with-recursive-descent/",
			Adapted:     true,
			Grammar: grammarFrom(`
				Regex  -> Term Regex'
				Regex' -> + Regex | ε
				Term   -> Factor Term | ε
				Factor -> Base Star
				Star   -> * Star | ε
				Base   -> id | ( Regex )`),
			Examples:        []string{"a b * + c", "( a + b ) * a b b", "a * *"},
			InvalidExamples: []string{"( a + b", "* a"},
		},
		{
			ID:              "dyck",
			Name:            "Parênteses balanceados (Dyck)",
			Description:     "A linguagem de Dyck: pares de ( ) e [ ] bem aninhados. A gramática do verbete já é LL(1); aqui ela foi estendida para dois tipos de par.",
			Group:           GroupLiterature,
			Author:          "Walther von Dyck (conceito); verbete da Wikipédia",
			Source:          "Dyck language, Wikipédia",
			URL:             "https://en.wikipedia.org/wiki/Dyck_language",
			Grammar:         grammarFrom(`S -> ( S ) S | [ S ] S | ε`),
			Examples:        []string{"( )", "[ ( ) ] ( )", "( ( [ ] ) )"},
			InvalidExamples: []string{"( ]", ") ("},
		},
		{
			ID:          "xml",
			Name:        "Elementos XML simplificados",
			Description: "Elementos com atributos, texto e elementos filhos, inclusive a tag vazia < br / >. As tags foram fatoradas à esquerda; a gramática não confere se a tag de fechamento tem o mesmo nome da de abertura (isso não cabe em uma gramática livre de contexto).",
			Group:       GroupLiterature,
			Author:      "Tim Bray, Jean Paoli, C. M. Sperberg-McQueen, Eve Maler, François Yergeau (eds.)",
			Source:      "Extensible Markup Language (XML) 1.0 (Fifth Edition), W3C Recommendation, 2008",
			URL:         "https://www.w3.org/TR/xml/",
			Adapted:     true,
			Grammar: grammarFrom(`
				Element -> < id Attrs Tail
				Attrs   -> id = id Attrs | ε
				Tail    -> / > | > Content
				Content -> id Content | < After
				After   -> id Attrs Tail Content | / id >`),
			Examples:        []string{"< br / >", "< p > ola mundo < / p >", "< a href = x > < b > t < / b > < / a >"},
			InvalidExamples: []string{"< p > texto", "< p x > < / p >"},
		},
		{
			ID:          "csv",
			Name:        "Registros CSV",
			Description: "Registros separados por ; (no lugar da quebra de linha) e campos separados por vírgula. Um campo pode ser vazio.",
			Group:       GroupLiterature,
			Author:      "Yakov Shafranovich",
			Source:      "RFC 4180: Common Format and MIME Type for Comma-Separated Values (CSV) Files, 2005",
			URL:         "https://www.rfc-editor.org/rfc/rfc4180",
			Adapted:     true,
			Grammar: grammarFrom(`
				File    -> Record File'
				File'   -> ; Record File' | ε
				Record  -> Field Record'
				Record' -> , Field Record' | ε
				Field   -> id | num | ε`),
			Examples:        []string{"nome , idade ; ana , 30 ; bia , 25", "a , , c", "1 ; 2 ; 3 ;"},
			InvalidExamples: []string{"a b , c", "a ; b c"},
		},
		{
			ID:          "logica-proposicional",
			Name:        "Fórmulas da lógica proposicional",
			Description: "Fórmulas bem formadas com ~ (negação), & (e), + (ou), > (implica) e = (bicondicional). Os parênteses das fórmulas binárias são obrigatórios.",
			Group:       GroupLiterature,
			Author:      "Kevin C. Klement",
			Source:      "Propositional Logic, Internet Encyclopedia of Philosophy",
			URL:         "https://iep.utm.edu/propositional-logic-sentential-logic/",
			Adapted:     true,
			Grammar: grammarFrom(`
				F  -> id | ~ F | ( F Op F )
				Op -> & | + | > | =`),
			Examples:        []string{"p", "~ ( p & q )", "( ( p > q ) = ( ~ q > ~ p ) )"},
			InvalidExamples: []string{"p & q", "( p & q & r )"},
		},
		{
			ID:          "lambda",
			Name:        "Cálculo lambda",
			Description: "Variáveis, abstrações \\ x [ corpo ] e aplicação por justaposição, associativa à esquerda. O λ virou \\ e a recursão à esquerda da aplicação foi removida.",
			Group:       GroupLiterature,
			Author:      "Alonzo Church (cálculo); Jesse Alama e Johannes Korbmacher (verbete)",
			Source:      "The Lambda Calculus, Stanford Encyclopedia of Philosophy, 2012, revisado em 2023",
			URL:         "https://plato.stanford.edu/entries/lambda-calculus/",
			Adapted:     true,
			Grammar: grammarFrom(`
				Term  -> Atom Term'
				Term' -> Atom Term' | ε
				Atom  -> id | ( Term ) | \ id [ Term ]`),
			Examples:        []string{"x", "\\ x [ x ]", "( \\ x [ x x ] ) ( \\ y [ y ] )"},
			InvalidExamples: []string{"\\ [ x ]", "( )"},
		},
		{
			ID:          "lua-tabelas",
			Name:        "Construtores de tabela Lua",
			Description: "Tabelas Lua com campos posicionais, nome = valor e [ chave ] = valor, separados por , ou ; e com separador final opcional. Name = exp e exp foram fatorados à esquerda.",
			Group:       GroupLiterature,
			Author:      "Roberto Ierusalimschy, Luiz Henrique de Figueiredo, Waldemar Celes",
			Source:      "Lua 5.4 Reference Manual, §3.4.9, PUC-Rio, 2020",
			URL:         "https://www.lua.org/manual/5.4/manual.html",
			Adapted:     true,
			Grammar: grammarFrom(`
				Table   -> { Fields }
				Fields  -> Field Fields' | ε
				Fields' -> Sep Fields | ε
				Sep     -> , | ;
				Field   -> [ Exp ] = Exp | id Field' | num | Table
				Field'  -> = Exp | ε
				Exp     -> id | num | Table`),
			Examples:        []string{"{ }", "{ 1 , 2 ; 3 , }", "{ x = 1 , [ k ] = { y } , z }"},
			InvalidExamples: []string{"{ , }", "{ x = }"},
		},
		{
			ID:          "wirth-ebnf",
			Name:        "Notação sintática de Wirth (EBNF)",
			Description: "A notação de Wirth para escrever gramáticas: produções nome = expressão . com opcionais [ ], grupos ( ), repetições { } e alternativas. A alternativa | virou /, como na ABNF.",
			Group:       GroupLiterature,
			Author:      "Niklaus Wirth",
			Source:      "What Can We Do about the Unnecessary Diversity of Notation for Syntactic Definitions?, Communications of the ACM 20(11), 1977",
			URL:         "https://www.cs.tufts.edu/comp/150FP/archive/niklaus-wirth/ebnf.pdf",
			Adapted:     true,
			Grammar: grammarFrom(`
				Syntax      -> Production Syntax | ε
				Production  -> id = Expression .
				Expression  -> Term Expression'
				Expression' -> / Term Expression' | ε
				Term        -> Factor Term'
				Term'       -> Factor Term' | ε
				Factor      -> id | [ Expression ] | ( Expression ) | { Expression }`),
			Examples:        []string{"s = a b .", "lista = id { virgula id } .", "a = [ b ] ( c / d ) . e = f ."},
			InvalidExamples: []string{"a = .", "a = b"},
		},
		{
			ID:          "brainfuck",
			Name:        "Brainf*ck",
			Description: "Os oito comandos > < + - . , [ ] com colchetes casados. Não há gramática publicada: esta foi escrita a partir da definição. Comentários não são aceitos.",
			Group:       GroupLiterature,
			Author:      "Urban Müller",
			Source:      "Brainfuck, Wikipédia (linguagem de 1993)",
			URL:         "https://en.wikipedia.org/wiki/Brainfuck",
			Grammar: grammarFrom(`
				Program -> Cmd Program | ε
				Cmd     -> + | - | < | > | . | , | [ Program ]`),
			Examples:        []string{"+ + [ - > + < ] .", ", [ . , ]", "> > + < -"},
			InvalidExamples: []string{"[ +", "+ ] ["},
		},
		{
			ID:          "pascal-tipos",
			Name:        "Declarações de variáveis Pascal",
			Description: "A seção var do Pascal com tipos nomeados, subintervalos 1 . . 10, ponteiros ^, arrays e records. .. virou . . e o array tem um índice só.",
			Group:       GroupLiterature,
			Author:      "ISO/IEC JTC1/SC22 (linguagem de Niklaus Wirth)",
			Source:      "ISO/IEC 7185:1990, Programming languages: Pascal, §6.2.1, §6.4 e §6.5.1",
			URL:         "https://www.standardpascal.org/iso7185.html",
			Adapted:     true,
			Grammar: grammarFrom(`
				VarPart -> var Decl ; Decls
				Decls   -> Decl ; Decls | ε
				Decl    -> id Ids : Type
				Ids     -> , id Ids | ε
				Type    -> id | Range | ^ id | array [ Range ] of Type | record Decl Rest end
				Range   -> num . . num
				Rest    -> ; Rest' | ε
				Rest'   -> Decl Rest | ε`),
			Examples: []string{
				"var x , y : integer ;",
				"var v : array [ 1 . . 10 ] of real ; p : ^ no ;",
				"var r : record a , b : integer ; c : 0 . . 9 ; end ;",
			},
			InvalidExamples: []string{"var : integer ;", "var a : array [ 1 . . 10 ] integer ;"},
		},
	}
}

// PresetByID acha uma gramática pronta pelo id.
func PresetByID(id string) (Preset, bool) {
	for _, p := range Presets() {
		if p.ID == id {
			return p, true
		}
	}
	return Preset{}, false
}

// alternativeSeparator separa as alternativas de uma regra em grammarFrom.
const alternativeSeparator = "|"

// grammarFrom lê uma regra por linha, "A -> α | β", com os símbolos separados
// por espaço. As cabeças, em ordem, são os não terminais (a primeira é o
// símbolo inicial); todo outro símbolo é terminal, na ordem em que aparece.
// As produções são numeradas na ordem de leitura.
func grammarFrom(rules string) *Grammar {
	g := &Grammar{}
	type rule struct{ head, body string }
	var parsed []rule
	for _, line := range strings.Split(rules, "\n") {
		head, body, found := strings.Cut(line, strings.TrimSpace(Arrow))
		if !found {
			continue
		}
		head = strings.TrimSpace(head)
		g.NonTerminals = append(g.NonTerminals, head)
		parsed = append(parsed, rule{head, body})
	}
	g.Start = g.NonTerminals[0]
	for _, r := range parsed {
		for _, alternative := range strings.Split(r.body, alternativeSeparator) {
			body := withoutInnerEpsilon(strings.Fields(alternative))
			for _, symbol := range body {
				if symbol != Epsilon && !g.IsNonTerminal(symbol) && !g.IsTerminal(symbol) {
					g.Terminals = append(g.Terminals, symbol)
				}
			}
			g.Productions = append(g.Productions, Production{Number: len(g.Productions) + 1, Head: r.head, Body: body})
		}
	}
	return g
}

// withoutInnerEpsilon tira o ε de um corpo com outros símbolos ("ε b" é só
// "b"); ε só fica quando é o corpo inteiro, senão o analisador o empilharia.
func withoutInnerEpsilon(body []string) []string {
	if len(body) <= 1 {
		return body
	}
	var result []string
	for _, symbol := range body {
		if symbol != Epsilon {
			result = append(result, symbol)
		}
	}
	if len(result) == 0 {
		return []string{Epsilon}
	}
	return result
}
