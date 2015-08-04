// A general lexer/parser for transforms.
// Generate using:
// go tool yacc -o pipeline_generator.go -p Transform transform_parser.y

%{
package transforms

import (
	//"fmt"
	"regexp"
	"strconv"
	"errors"
)
%}

// fields inside this union end up as the fields in a structure known
// as ${PREFIX}SymType, of which a reference is passed to the lexer.

%union{
	val TransformFunc
	strVal string
}

// All transforms return a TransformFunc
%type <val> or_test and_test not_test comparison terminal if_transform transform_list

// All tokens and terminals are strings
%token <strVal> NUMBER BOOL STRING COMPOP GET OR AND NOT RB LB HAS EOF IF PIPE

%%

transform_list
	: if_transform
		{
			Transformlex.(*TransformLex).output = $1
		}
	| transform_list PIPE if_transform
		{
			$$ = pipelineGeneratorTransform($1, $3)
			Transformlex.(*TransformLex).output = $$
		}
	;

if_transform
	: or_test
	| IF or_test
		{
			$$ = pipelineGeneratorIf($2)
		}
	;


or_test
    : and_test
    | or_test OR and_test
		{
			$$ = pipelineGeneratorOr($1, $3)
		}
	;

and_test
    : not_test
    | and_test AND not_test
		{
			$$ = pipelineGeneratorAnd($1, $3)
		}
	;

not_test
    : comparison
    | NOT not_test
		{
			$$ = pipelineGeneratorNot($2)
		}
    ;

comparison:
	terminal
	| terminal COMPOP terminal
		{
			$$ = pipelineGeneratorCompare($1, $3, $2)
		}
    ;

terminal:
    NUMBER
		{
			num, err := strconv.ParseFloat($1, 64)
			$$ = pipelineGeneratorConstant(num, err)
		}
    | BOOL
		{
			val, err := strconv.ParseBool($1)
			$$ = pipelineGeneratorConstant(val, err)
		}
    | GET RB
		{
			$$ = pipelineGeneratorIdentity()
		}
    | GET STRING RB
		{
			$$ = pipelineGeneratorGet($2)
		}
    | HAS STRING RB
		{
			$$ = pipelineGeneratorHas($2)
		}
	| STRING
		{
			$$ = pipelineGeneratorConstant($1, nil)
		}
	| LB or_test RB
		{
			$$ = $2
		}
    ;

%%  /* Start of lexer, hopefully go will let us do this automatically in the future */


const (
	eof = 0
	errorString = "<ERROR>"
	eofString = "<EOF>"
)

var tokenizer *regexp.Regexp

func init() {
	tk, err := regexp.Compile(`^((get|has)\(|(-)?[0-9]+(.[0-9]+)?|\".+?\"|\)|\(|true|false|and|or|not|(<=|>=|<|>|==|!=)|:|if)`)
	if err != nil {
		panic(err.Error())
	}
	tokenizer = tk
}


// ParseTransform takes a transform input and returns a function to do the
// transforms.
func ParseTransform(input string) (TransformFunc, error) {
	tl := TransformLex{input:input}

	TransformParse(&tl)

	if tl.errorString == "" {
		return tl.output, nil
	}

	return tl.output, errors.New(tl.errorString)
}

type TransformLex struct {
	input string
	position int

	errorString string
	output TransformFunc
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
	case "get(":
		return GET
	case "has(":
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
	case ":":
		return PIPE
	default:
		if token[0] == '"' || token[0] == '\'' {

			// unquote token
			lval.strVal = token[1: len(token) - 1]

			return STRING
		}

		return NUMBER
	}
}

func (l *TransformLex) Error(s string) {
	l.errorString = s
}
