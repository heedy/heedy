%{

package transforms

import (
	"fmt"
	"regexp"
	"strconv"
)

%}

// fields inside this union end up as the fields in a structure known
// as ${PREFIX}SymType, of which a reference is passed to the lexer.

%union{
	val TransformFunc
	strVal string
}

%type <val> transform and_test not_test comparison terminal transformlist
%token <strVal> NUMBER BOOL STRING COMPOP


transformlist:
    : transform EOF
        {
            return $1;
        }
    ;


transform
    : and_test
    | or_test "or" and_test {
		$$ = pipelineGeneratorOr($1, $3)
	}

and_test:
    not_test
    | and_test "and" not_test {
		$$ = pipelineGeneratorAnd($1, $3)
	}

not_test:
    comparison
    | "not" not_test {
		$$ = pipelineGeneratorNot($2)
	}
    ;

comparison:
	terminal
	| terminal COMPOP terminal {
		$$ = pipelineGeneratorCompare($1, $3, $2)
	}
    ;

terminal:
    NUMBER {
		num, err := strconv.ParseFloat($1, 64)
		$$ = pipelineGeneratorConstant(num, err)
	}
    | BOOL {
		val, err := strconv.ParseBool($1, 64)
		$$ = pipelineGeneratorConstant(val, err)
	}
    | "get(" ")" {
		$$ = pipelineGeneratorIdentity()
	}
    | "get(" STRING ")" {
		$$ = pipelineGeneratorGet($2)
	}
    | "has(" STRING ")" {
		$$ = pipelineGeneratorHas($2)
	}
	| STRING {
		$$ = pipelineGeneratorConstant($1, nil)
	}
	| "(" and_test ")" {
		return $2
	}
    ;

%%  /* Start of lexer (why the heck do we have to generate this by hand?) */


var tokenizer regexp.Regexp

func init() {
	tokenizer, _ = regexp.Compile(`^((get|has)\(|(-)?[0-9]+(.[0-9]+)?|\".+?\"|\)|\(|true|false|and|or|not|(<=|>=|<|>|==|!=)|:|if)`)
}


const (
	eof = 0
	errorString = "<ERROR>"
)

type TransformLex struct {
	input string
	position int

	errorString string
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
			return ""
		}
		c = rune(l.input[l.position])
		l.position += 1
	}

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
	fmt.Println(token)
	lval.strVal = token

	switch token {
	case "":
		return eof
	case errorString:
		return eof
	case "true", "false":
		return BOOL
	case ")":
		return int(')')
	case "get(":
		return int('g')
	case "has(":
		return int('h')
	case "and":
		return int('a')
	case "or":
		return int('o')
	case "not":
		return int('n')
	case ">=", "<=", ">", "<", "==", "!=":
		return int('=')
	default:
		_, err := func strconv.ParseFloat(token, 64)
		if err != nil {
			return NUMBER
		}

		return STRING
	}
}

func (l *TransformLex) Error(s string) {
	l.errorString = s
}


func main() {
	fi := bufio.NewReader(os.NewFile(0, "stdin"))

	for {
		var eqn string
		var ok bool

		fmt.Printf("equation: ")
		if eqn, ok = readline(fi); ok {
			CalcParse(&CalcLex{s: eqn})
		} else {
			break
		}
	}
}

func readline(fi *bufio.Reader) (string, bool) {
	s, err := fi.ReadString('\n')
	if err != nil {
		return "", false
	}
	return s, true
}
