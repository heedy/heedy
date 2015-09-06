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
	"strings"
)
%}

// fields inside this union end up as the fields in a structure known
// as ${PREFIX}SymType, of which a reference is passed to the lexer.

%union{
	val TransformFunc
	strVal string
	stringList []string
	funcList   []TransformFunc
}

// All transforms return a TransformFunc
%type <val> or_test and_test not_test comparison terminal if_transform transform_list constant variable function term expression
%type <funcList> function_params transform_list_params
%type <stringList> string_list

// All tokens and terminals are strings
%token <strVal> NUMBER BOOL STRING COMPOP THIS OR AND NOT RB LB EOF PIPE RSQUARE LSQUARE COMMA IDENTIFIER HAS IF SET PLUS MINUS MULTIPLY DIVIDE

%left UMINUS      /*  supplies  precedence  for  unary  minus  */


%%

transform_list
    : transform_list_params
	    {
			$$ = pipeline($1)
			Transformlex.(*TransformLex).output = $$
		}
	;

transform_list_params
	: if_transform
		{
			$$ = []TransformFunc{$1}
		}
	| transform_list_params PIPE if_transform
		{
			//$$ = append([]TransformFunc{$3}, $1...)
			$$ = append($1, $3)
		}
	;


if_transform
	: or_test
	| IF or_test
		{
			$$ = pipelineGeneratorIf($2)
		}
	| IDENTIFIER
		{
			fun, err := InstantiateRegisteredFunction($1)

			if err != nil {
				Transformlex.Error(err.Error())
			}

			$$ = fun
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

comparison
	: expression
	| expression COMPOP expression
		{
			$$ = pipelineGeneratorCompare($1, $3, $2)
		}
    ;

expression
	: term
	| expression PLUS term
		{
			$$ = addTransformGenerator($1, $3)
		}
	| expression MINUS term
		{
			$$ = subtractTransformGenerator($1, $3)
		}
	;

term
	: terminal
	| term MULTIPLY terminal
		{
			$$ = multiplyTransformGenerator($1, $3)
		}
	| term DIVIDE terminal
		{
			$$ = divideTransformGenerator($1, $3)
		}
	| MINUS terminal %prec  UMINUS
		{
			$$ = inverseTransformGenerator($2)
		}
	;

terminal
	: constant
	| variable
	| function
	| LB transform_list RB
		{
			$$ = $2
		}
	;

constant
	: NUMBER
		{
			num, err := strconv.ParseFloat($1, 64)
			$$ = ConstantValueGenerator(num, err)
		}
    | BOOL
		{
			val, err := strconv.ParseBool($1)
			$$ = ConstantValueGenerator(val, err)
		}
    | STRING
		{
			$$ = ConstantValueGenerator($1, nil)
		}
    ;

variable
	: THIS LSQUARE string_list RSQUARE
		{
			$$ = pipelineGeneratorGet($3)
		}
	| THIS
		{
			$$ = PipelineGeneratorIdentity()
		}
	;

function
	: SET LB THIS LSQUARE string_list RSQUARE COMMA or_test RB
		{
			$$ = pipelineGeneratorSet($5, $8)
		}
	| SET LB THIS COMMA transform_list RB
		{
			$$ = pipelineGeneratorSet([]string{}, $5)
		}
	| HAS LB STRING RB
		{
			$$ = pipelineGeneratorHas($3)
		}
	| IDENTIFIER LB RB
		{
			fun, err := InstantiateRegisteredFunction($1)

			if err != nil {
				Transformlex.Error(err.Error())
			}

			$$ = fun
		}
	| IDENTIFIER LB function_params RB
		{
			fun, err := InstantiateRegisteredFunction($1, $3...)

			if err != nil {
				Transformlex.Error(err.Error())
			}

			$$ = fun
		}
	;

string_list
	: STRING
		{
			$$ = []string{$1}
		}
	| string_list COMMA STRING
		{
			$$ = append($1, $3)
		}
	;

function_params
	: transform_list
		{
			$$ = []TransformFunc{$1}
		}
	| function_params COMMA transform_list
		{
			$$ = append([]TransformFunc{$3}, $1...)
		}
	;

%%  /* Start of lexer, hopefully go will let us do this automatically in the future */


const (
	eof = 0
	errorString = "<ERROR>"
	eofString = "<EOF>"
	builtins = `has|if|set`
	logicals  = `true|false|and|or|not`
	numbers   = `(-)?[0-9]+(\.[0-9]+)?`
	compops   = `<=|>=|<|>|==|!=`
	stringr   = `\"(\\["nrt\\]|.)*?\"|'(\\['nrt\\]|.)*?'`
	pipes     = `:|\||,`
	syms      = `\$|\[|\]|\(|\)`
	idents    = `([a-zA-Z_][a-zA-Z_0-9]*)`
	maths     = `\-|\*|/|\+`
	allregex = builtins  + "|" + logicals  + "|" + numbers  + "|" + compops + "|" + stringr + "|" + pipes  + "|" + syms  + "|" + idents + "|" + maths

)

var (
	tokenizer   = regexp.MustCompile(`^(` + allregex + `)`)
	numberRegex = regexp.MustCompile("^" + numbers + "$")
	stringRegex = regexp.MustCompile("^" + stringr + "$")
	identRegex  = regexp.MustCompile("^" + idents + "$")
)

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
		lexer.Error("Error, unknown token")
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
				strval := token[1: len(token) - 1]

				// replace escape characters
				strval = strings.Replace(strval, "\\n", "\n", -1)
				strval = strings.Replace(strval, "\\r", "\r", -1)
				strval = strings.Replace(strval, "\\t", "\t", -1)
				strval = strings.Replace(strval, "\\\\", "\\", -1)
				strval = strings.Replace(strval, "\\\"", "\"", -1)
				strval = strings.Replace(strval, "\\'", "'", -1)

				lval.strVal = strval
				return STRING
			default:
				return IDENTIFIER
		}
	}
}

func (l *TransformLex) Error(s string) {
	l.errorString = s
}
