//line pipeline_generator.y:6
package transforms

import __yyfmt__ "fmt"

//line pipeline_generator.y:6
import (
	//"fmt"
	"errors"
	"regexp"
	"strconv"
	"strings"
)

//line pipeline_generator.y:20
type TransformSymType struct {
	yys        int
	val        TransformFunc
	strVal     string
	stringList []string
	funcList   []TransformFunc
}

const NUMBER = 57346
const BOOL = 57347
const STRING = 57348
const COMPOP = 57349
const THIS = 57350
const OR = 57351
const AND = 57352
const NOT = 57353
const RB = 57354
const LB = 57355
const EOF = 57356
const PIPE = 57357
const RSQUARE = 57358
const LSQUARE = 57359
const COMMA = 57360
const GTE = 57361
const LTE = 57362
const GT = 57363
const LT = 57364
const EQ = 57365
const NE = 57366
const IDENTIFIER = 57367
const HAS = 57368
const IF = 57369
const SET = 57370
const PLUS = 57371
const MINUS = 57372
const MULTIPLY = 57373
const DIVIDE = 57374

var TransformToknames = []string{
	"NUMBER",
	"BOOL",
	"STRING",
	"COMPOP",
	"THIS",
	"OR",
	"AND",
	"NOT",
	"RB",
	"LB",
	"EOF",
	"PIPE",
	"RSQUARE",
	"LSQUARE",
	"COMMA",
	"GTE",
	"LTE",
	"GT",
	"LT",
	"EQ",
	"NE",
	"IDENTIFIER",
	"HAS",
	"IF",
	"SET",
	"PLUS",
	"MINUS",
	"MULTIPLY",
	"DIVIDE",
}
var TransformStatenames = []string{}

const TransformEofCode = 1
const TransformErrCode = 2
const TransformMaxDepth = 200

//line pipeline_generator.y:240

/* Start of lexer, hopefully go will let us do this automatically in the future */

const (
	eof         = 0
	errorString = "<ERROR>"
	eofString   = "<EOF>"
	builtins    = `has|if|gte|lte|gt|lt|eq|ne|set`
	logicals    = `true|false|and|or|not`
	numbers     = `(-)?[0-9]+(\.[0-9]+)?`
	compops     = `<=|>=|<|>|==|!=`
	stringr     = `\".+?\"`
	pipes       = `:|\||,`
	syms        = `\$|\[|\]|\(|\)`
	idents      = `([a-zA-Z_][a-zA-Z_0-9]*)`
	maths       = `\-|\*|/|\+`
)

var (
	tokenizer   *regexp.Regexp
	numberRegex *regexp.Regexp
	stringRegex *regexp.Regexp
	identRegex  *regexp.Regexp
)

func init() {

	var err error
	{
		re := strings.Join([]string{builtins, logicals, numbers, compops, stringr, pipes, syms, idents, maths}, "|")

		regexStr := `^(` + re + `)`
		tokenizer, err = regexp.Compile(regexStr)
		if err != nil {
			panic(err.Error())
		}
	}

	// these regexes are needed later on while testing.
	numberRegex, err = regexp.Compile("^" + numbers + "$")
	if err != nil {
		panic(err.Error())
	}

	// string regex (needed later on)
	stringRegex, err = regexp.Compile("^" + stringr + "$")
	if err != nil {
		panic(err.Error())
	}

	// ident regex
	identRegex, err = regexp.Compile("^" + idents + "$")
	if err != nil {
		panic(err.Error())
	}
}

// ParseTransform takes a transform input and returns a function to do the
// transforms.
func ParseTransform(input string) (TransformFunc, error) {
	tl := TransformLex{input: input}

	TransformParse(&tl)

	if tl.errorString == "" {
		return tl.output, nil
	}

	return tl.output, errors.New(tl.errorString)
}

type TransformLex struct {
	input    string
	position int

	errorString string
	output      TransformFunc
}

// Are we at the end of file?
func (t *TransformLex) AtEOF() bool {
	return t.position >= len(t.input)
}

// Return the next string for the lexer
func (l *TransformLex) Next() string {
	var c rune = ' '

	// skip whitespace
	for c == ' ' || c == '\t' {
		if l.AtEOF() {
			return eofString
		}
		c = rune(l.input[l.position])
		l.position += 1
	}

	l.position -= 1

	rest := l.input[l.position:]

	token := tokenizer.FindString(rest)
	l.position += len(token)

	if token == "" {
		return errorString
	}

	return token
}

func (lexer *TransformLex) Lex(lval *TransformSymType) int {

	token := lexer.Next()
	//fmt.Println("token: " + token)
	lval.strVal = token

	switch token {
	case eofString:
		return 0
	case errorString:
		return 0
	case "true", "false":
		return BOOL
	case ")":
		return RB
	case "(":
		return LB
	case "[":
		return LSQUARE
	case "]":
		return RSQUARE
	case "$":
		return THIS
	case "has":
		return HAS
	case "and":
		return AND
	case "or":
		return OR
	case "not":
		return NOT
	case ">=", "<=", ">", "<", "==", "!=":
		return COMPOP
	case "if":
		return IF
	case "|", ":":
		return PIPE
	case ",":
		return COMMA
	case "gte":
		return GTE
	case "lte":
		return LTE
	case "lt":
		return LT
	case "gt":
		return GT
	case "eq":
		return EQ
	case "ne":
		return NE
	case "set":
		return SET
	case "-":
		return MINUS
	case "+":
		return PLUS
	case "/":
		return DIVIDE
	case "*":
		return MULTIPLY
	default:
		switch {
		case numberRegex.MatchString(token):
			return NUMBER
		case stringRegex.MatchString(token):
			// unquote token
			lval.strVal = token[1 : len(token)-1]

			return STRING
		default:
			return IDENTIFIER
		}
	}
}

func (l *TransformLex) Error(s string) {
	l.errorString = s
}

//line yacctab:1
var TransformExca = []int{
	-1, 1,
	1, -1,
	-2, 0,
}

const TransformNprod = 43
const TransformPrivate = 57344

var TransformTokenNames []string
var TransformStates []string

const TransformLast = 150

var TransformAct = []int{

	1, 3, 59, 16, 17, 18, 31, 19, 9, 34,
	8, 69, 15, 37, 38, 91, 39, 40, 22, 23,
	24, 25, 26, 27, 28, 21, 4, 20, 10, 6,
	5, 35, 36, 35, 36, 89, 83, 73, 33, 74,
	75, 29, 84, 53, 63, 64, 65, 66, 67, 68,
	71, 16, 17, 18, 2, 19, 49, 72, 8, 73,
	15, 51, 52, 48, 54, 55, 22, 23, 24, 25,
	26, 27, 28, 21, 4, 20, 87, 86, 11, 16,
	17, 18, 47, 19, 50, 88, 8, 30, 15, 90,
	93, 76, 29, 92, 22, 23, 24, 25, 26, 27,
	28, 21, 46, 20, 16, 17, 18, 82, 19, 81,
	29, 80, 29, 15, 29, 79, 56, 57, 29, 22,
	23, 24, 25, 26, 27, 28, 21, 78, 20, 77,
	29, 58, 29, 45, 29, 44, 43, 42, 41, 32,
	30, 61, 60, 85, 62, 70, 14, 13, 12, 7,
}
var TransformPact = []int{

	47, 26, -1000, 131, 75, 129, -1000, -1000, 75, 2,
	-18, -1000, -1000, -1000, -1000, 47, -1000, -1000, -1000, 0,
	125, 124, 123, 122, 120, 89, 69, 50, 43, 47,
	75, 131, 75, -1000, 100, 100, 100, 100, 100, 119,
	136, 133, 138, 47, 47, 47, 47, 47, 47, -1,
	-1000, 129, -1000, 4, -18, -18, -1000, -1000, -1000, 41,
	-1000, 22, 79, 117, 115, 103, 99, 97, 95, -1000,
	24, 26, -1000, 137, 136, 47, -1000, -1000, -1000, -1000,
	-1000, -1000, -1000, -1000, 47, -1000, 19, 77, 26, -3,
	-1000, 75, 78, -1000,
}
var TransformPgo = []int{

	0, 1, 30, 29, 149, 78, 54, 0, 148, 147,
	146, 28, 8, 145, 2,
}
var TransformR1 = []int{

	0, 7, 7, 6, 6, 1, 1, 2, 2, 3,
	3, 4, 4, 12, 12, 12, 11, 11, 11, 5,
	5, 5, 5, 8, 8, 8, 9, 9, 10, 10,
	10, 10, 10, 10, 10, 10, 10, 10, 10, 14,
	14, 13, 13,
}
var TransformR2 = []int{

	0, 1, 3, 1, 2, 1, 3, 1, 3, 1,
	2, 1, 3, 1, 3, 3, 1, 3, 3, 1,
	1, 1, 3, 1, 1, 1, 4, 1, 9, 6,
	4, 4, 4, 4, 4, 4, 4, 3, 4, 1,
	3, 1, 3,
}
var TransformChk = []int{

	-1000, -7, -6, -1, 27, -2, -3, -4, 11, -12,
	-11, -5, -8, -9, -10, 13, 4, 5, 6, 8,
	28, 26, 19, 20, 21, 22, 23, 24, 25, 15,
	9, -1, 10, -3, 7, 29, 30, 31, 32, -7,
	17, 13, 13, 13, 13, 13, 13, 13, 13, 13,
	-6, -2, -3, -12, -11, -11, -5, -5, 12, -14,
	6, 8, 6, -7, -7, -7, -7, -7, -7, 12,
	-13, -7, 16, 18, 17, 18, 12, 12, 12, 12,
	12, 12, 12, 12, 18, 6, -14, -7, -7, 16,
	12, 18, -1, 12,
}
var TransformDef = []int{

	0, -2, 1, 3, 0, 5, 7, 9, 0, 11,
	13, 16, 19, 20, 21, 0, 23, 24, 25, 27,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 4, 0, 10, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	2, 6, 8, 12, 14, 15, 17, 18, 22, 0,
	39, 0, 0, 0, 0, 0, 0, 0, 0, 37,
	0, 41, 26, 0, 0, 0, 30, 31, 32, 33,
	34, 35, 36, 38, 0, 40, 0, 0, 42, 0,
	29, 0, 0, 28,
}
var TransformTok1 = []int{

	1,
}
var TransformTok2 = []int{

	2, 3, 4, 5, 6, 7, 8, 9, 10, 11,
	12, 13, 14, 15, 16, 17, 18, 19, 20, 21,
	22, 23, 24, 25, 26, 27, 28, 29, 30, 31,
	32,
}
var TransformTok3 = []int{
	0,
}

//line yaccpar:1

/*	parser for yacc output	*/

var TransformDebug = 0

type TransformLexer interface {
	Lex(lval *TransformSymType) int
	Error(s string)
}

const TransformFlag = -1000

func TransformTokname(c int) string {
	// 4 is TOKSTART above
	if c >= 4 && c-4 < len(TransformToknames) {
		if TransformToknames[c-4] != "" {
			return TransformToknames[c-4]
		}
	}
	return __yyfmt__.Sprintf("tok-%v", c)
}

func TransformStatname(s int) string {
	if s >= 0 && s < len(TransformStatenames) {
		if TransformStatenames[s] != "" {
			return TransformStatenames[s]
		}
	}
	return __yyfmt__.Sprintf("state-%v", s)
}

func Transformlex1(lex TransformLexer, lval *TransformSymType) int {
	c := 0
	char := lex.Lex(lval)
	if char <= 0 {
		c = TransformTok1[0]
		goto out
	}
	if char < len(TransformTok1) {
		c = TransformTok1[char]
		goto out
	}
	if char >= TransformPrivate {
		if char < TransformPrivate+len(TransformTok2) {
			c = TransformTok2[char-TransformPrivate]
			goto out
		}
	}
	for i := 0; i < len(TransformTok3); i += 2 {
		c = TransformTok3[i+0]
		if c == char {
			c = TransformTok3[i+1]
			goto out
		}
	}

out:
	if c == 0 {
		c = TransformTok2[1] /* unknown char */
	}
	if TransformDebug >= 3 {
		__yyfmt__.Printf("lex %s(%d)\n", TransformTokname(c), uint(char))
	}
	return c
}

func TransformParse(Transformlex TransformLexer) int {
	var Transformn int
	var Transformlval TransformSymType
	var TransformVAL TransformSymType
	TransformS := make([]TransformSymType, TransformMaxDepth)

	Nerrs := 0   /* number of errors */
	Errflag := 0 /* error recovery flag */
	Transformstate := 0
	Transformchar := -1
	Transformp := -1
	goto Transformstack

ret0:
	return 0

ret1:
	return 1

Transformstack:
	/* put a state and value onto the stack */
	if TransformDebug >= 4 {
		__yyfmt__.Printf("char %v in %v\n", TransformTokname(Transformchar), TransformStatname(Transformstate))
	}

	Transformp++
	if Transformp >= len(TransformS) {
		nyys := make([]TransformSymType, len(TransformS)*2)
		copy(nyys, TransformS)
		TransformS = nyys
	}
	TransformS[Transformp] = TransformVAL
	TransformS[Transformp].yys = Transformstate

Transformnewstate:
	Transformn = TransformPact[Transformstate]
	if Transformn <= TransformFlag {
		goto Transformdefault /* simple state */
	}
	if Transformchar < 0 {
		Transformchar = Transformlex1(Transformlex, &Transformlval)
	}
	Transformn += Transformchar
	if Transformn < 0 || Transformn >= TransformLast {
		goto Transformdefault
	}
	Transformn = TransformAct[Transformn]
	if TransformChk[Transformn] == Transformchar { /* valid shift */
		Transformchar = -1
		TransformVAL = Transformlval
		Transformstate = Transformn
		if Errflag > 0 {
			Errflag--
		}
		goto Transformstack
	}

Transformdefault:
	/* default state action */
	Transformn = TransformDef[Transformstate]
	if Transformn == -2 {
		if Transformchar < 0 {
			Transformchar = Transformlex1(Transformlex, &Transformlval)
		}

		/* look through exception table */
		xi := 0
		for {
			if TransformExca[xi+0] == -1 && TransformExca[xi+1] == Transformstate {
				break
			}
			xi += 2
		}
		for xi += 2; ; xi += 2 {
			Transformn = TransformExca[xi+0]
			if Transformn < 0 || Transformn == Transformchar {
				break
			}
		}
		Transformn = TransformExca[xi+1]
		if Transformn < 0 {
			goto ret0
		}
	}
	if Transformn == 0 {
		/* error ... attempt to resume parsing */
		switch Errflag {
		case 0: /* brand new error */
			Transformlex.Error("syntax error")
			Nerrs++
			if TransformDebug >= 1 {
				__yyfmt__.Printf("%s", TransformStatname(Transformstate))
				__yyfmt__.Printf(" saw %s\n", TransformTokname(Transformchar))
			}
			fallthrough

		case 1, 2: /* incompletely recovered error ... try again */
			Errflag = 3

			/* find a state where "error" is a legal shift action */
			for Transformp >= 0 {
				Transformn = TransformPact[TransformS[Transformp].yys] + TransformErrCode
				if Transformn >= 0 && Transformn < TransformLast {
					Transformstate = TransformAct[Transformn] /* simulate a shift of "error" */
					if TransformChk[Transformstate] == TransformErrCode {
						goto Transformstack
					}
				}

				/* the current p has no shift on "error", pop stack */
				if TransformDebug >= 2 {
					__yyfmt__.Printf("error recovery pops state %d\n", TransformS[Transformp].yys)
				}
				Transformp--
			}
			/* there is no state on the stack with an error shift ... abort */
			goto ret1

		case 3: /* no shift yet; clobber input char */
			if TransformDebug >= 2 {
				__yyfmt__.Printf("error recovery discards %s\n", TransformTokname(Transformchar))
			}
			if Transformchar == TransformEofCode {
				goto ret1
			}
			Transformchar = -1
			goto Transformnewstate /* try again in the same state */
		}
	}

	/* reduction by production Transformn */
	if TransformDebug >= 2 {
		__yyfmt__.Printf("reduce %v in:\n\t%v\n", Transformn, TransformStatname(Transformstate))
	}

	Transformnt := Transformn
	Transformpt := Transformp
	_ = Transformpt // guard against "declared and not used"

	Transformp -= TransformR2[Transformn]
	TransformVAL = TransformS[Transformp+1]

	/* consult goto table to find next state */
	Transformn = TransformR1[Transformn]
	Transformg := TransformPgo[Transformn]
	Transformj := Transformg + TransformS[Transformp].yys + 1

	if Transformj >= TransformLast {
		Transformstate = TransformAct[Transformg]
	} else {
		Transformstate = TransformAct[Transformj]
		if TransformChk[Transformstate] != -Transformn {
			Transformstate = TransformAct[Transformg]
		}
	}
	// dummy call; replaced with literal code
	switch Transformnt {

	case 1:
		//line pipeline_generator.y:39
		{
			Transformlex.(*TransformLex).output = TransformS[Transformpt-0].val
			TransformVAL.val = TransformS[Transformpt-0].val
		}
	case 2:
		//line pipeline_generator.y:44
		{
			TransformVAL.val = pipelineGeneratorTransform(TransformS[Transformpt-2].val, TransformS[Transformpt-0].val)
			Transformlex.(*TransformLex).output = TransformVAL.val
		}
	case 3:
		TransformVAL.val = TransformS[Transformpt-0].val
	case 4:
		//line pipeline_generator.y:53
		{
			TransformVAL.val = pipelineGeneratorIf(TransformS[Transformpt-0].val)
		}
	case 5:
		TransformVAL.val = TransformS[Transformpt-0].val
	case 6:
		//line pipeline_generator.y:62
		{
			TransformVAL.val = pipelineGeneratorOr(TransformS[Transformpt-2].val, TransformS[Transformpt-0].val)
		}
	case 7:
		TransformVAL.val = TransformS[Transformpt-0].val
	case 8:
		//line pipeline_generator.y:70
		{
			TransformVAL.val = pipelineGeneratorAnd(TransformS[Transformpt-2].val, TransformS[Transformpt-0].val)
		}
	case 9:
		TransformVAL.val = TransformS[Transformpt-0].val
	case 10:
		//line pipeline_generator.y:78
		{
			TransformVAL.val = pipelineGeneratorNot(TransformS[Transformpt-0].val)
		}
	case 11:
		TransformVAL.val = TransformS[Transformpt-0].val
	case 12:
		//line pipeline_generator.y:86
		{
			TransformVAL.val = pipelineGeneratorCompare(TransformS[Transformpt-2].val, TransformS[Transformpt-0].val, TransformS[Transformpt-1].strVal)
		}
	case 13:
		TransformVAL.val = TransformS[Transformpt-0].val
	case 14:
		//line pipeline_generator.y:94
		{
			TransformVAL.val = addTransformGenerator(TransformS[Transformpt-2].val, TransformS[Transformpt-0].val)
		}
	case 15:
		//line pipeline_generator.y:98
		{
			TransformVAL.val = subtractTransformGenerator(TransformS[Transformpt-2].val, TransformS[Transformpt-0].val)
		}
	case 16:
		TransformVAL.val = TransformS[Transformpt-0].val
	case 17:
		//line pipeline_generator.y:106
		{
			TransformVAL.val = multiplyTransformGenerator(TransformS[Transformpt-2].val, TransformS[Transformpt-0].val)
		}
	case 18:
		//line pipeline_generator.y:110
		{
			TransformVAL.val = divideTransformGenerator(TransformS[Transformpt-2].val, TransformS[Transformpt-0].val)
		}
	case 19:
		TransformVAL.val = TransformS[Transformpt-0].val
	case 20:
		TransformVAL.val = TransformS[Transformpt-0].val
	case 21:
		TransformVAL.val = TransformS[Transformpt-0].val
	case 22:
		//line pipeline_generator.y:120
		{
			TransformVAL.val = TransformS[Transformpt-1].val
		}
	case 23:
		//line pipeline_generator.y:127
		{
			num, err := strconv.ParseFloat(TransformS[Transformpt-0].strVal, 64)
			TransformVAL.val = pipelineGeneratorConstant(num, err)
		}
	case 24:
		//line pipeline_generator.y:132
		{
			val, err := strconv.ParseBool(TransformS[Transformpt-0].strVal)
			TransformVAL.val = pipelineGeneratorConstant(val, err)
		}
	case 25:
		//line pipeline_generator.y:137
		{
			TransformVAL.val = pipelineGeneratorConstant(TransformS[Transformpt-0].strVal, nil)
		}
	case 26:
		//line pipeline_generator.y:144
		{
			TransformVAL.val = pipelineGeneratorGet(TransformS[Transformpt-1].stringList)
		}
	case 27:
		//line pipeline_generator.y:148
		{
			TransformVAL.val = pipelineGeneratorIdentity()
		}
	case 28:
		//line pipeline_generator.y:155
		{
			TransformVAL.val = pipelineGeneratorSet(TransformS[Transformpt-4].stringList, TransformS[Transformpt-1].val)
		}
	case 29:
		//line pipeline_generator.y:159
		{
			TransformVAL.val = pipelineGeneratorSet([]string{}, TransformS[Transformpt-1].val)
		}
	case 30:
		//line pipeline_generator.y:163
		{
			TransformVAL.val = pipelineGeneratorHas(TransformS[Transformpt-1].strVal)
		}
	case 31:
		//line pipeline_generator.y:167
		{
			identity := pipelineGeneratorIdentity()
			TransformVAL.val = pipelineGeneratorCompare(identity, TransformS[Transformpt-1].val, ">=")
		}
	case 32:
		//line pipeline_generator.y:172
		{
			identity := pipelineGeneratorIdentity()
			TransformVAL.val = pipelineGeneratorCompare(identity, TransformS[Transformpt-1].val, "<=")
		}
	case 33:
		//line pipeline_generator.y:177
		{
			identity := pipelineGeneratorIdentity()
			TransformVAL.val = pipelineGeneratorCompare(identity, TransformS[Transformpt-1].val, ">")
		}
	case 34:
		//line pipeline_generator.y:182
		{
			identity := pipelineGeneratorIdentity()
			TransformVAL.val = pipelineGeneratorCompare(identity, TransformS[Transformpt-1].val, "<")
		}
	case 35:
		//line pipeline_generator.y:187
		{
			identity := pipelineGeneratorIdentity()
			TransformVAL.val = pipelineGeneratorCompare(identity, TransformS[Transformpt-1].val, "==")
		}
	case 36:
		//line pipeline_generator.y:192
		{
			identity := pipelineGeneratorIdentity()
			TransformVAL.val = pipelineGeneratorCompare(identity, TransformS[Transformpt-1].val, "!=")
		}
	case 37:
		//line pipeline_generator.y:197
		{
			fun, err := getCustomFunction(TransformS[Transformpt-2].strVal)

			if err != nil {
				Transformlex.Error(err.Error())
			}

			TransformVAL.val = fun
		}
	case 38:
		//line pipeline_generator.y:207
		{
			fun, err := getCustomFunction(TransformS[Transformpt-3].strVal, TransformS[Transformpt-1].funcList...)

			if err != nil {
				Transformlex.Error(err.Error())
			}

			TransformVAL.val = fun
		}
	case 39:
		//line pipeline_generator.y:220
		{
			TransformVAL.stringList = []string{TransformS[Transformpt-0].strVal}
		}
	case 40:
		//line pipeline_generator.y:224
		{
			TransformVAL.stringList = append([]string{TransformS[Transformpt-0].strVal}, TransformS[Transformpt-2].stringList...)
		}
	case 41:
		//line pipeline_generator.y:231
		{
			TransformVAL.funcList = []TransformFunc{TransformS[Transformpt-0].val}
		}
	case 42:
		//line pipeline_generator.y:235
		{
			TransformVAL.funcList = append([]TransformFunc{TransformS[Transformpt-0].val}, TransformS[Transformpt-2].funcList...)
		}
	}
	goto Transformstack /* stack new state and value */
}
