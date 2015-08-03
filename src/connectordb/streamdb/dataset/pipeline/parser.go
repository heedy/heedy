package pipeline

import (
	"errors"
)

/*
The pipeline Parser is a very simple method to parse things of the format:
blah:blah2(arg1,arg2):blah

There are several things here: this language does not define any variables, so everything is a string.
Therefore, this is valid:
"blah blah"(arg1):blah
and so is this
blah blah(arg1)
since a symbol goes on until a special character is recognized.
*/

const (
	pipeSymbol         = byte(':')
	startArgsSymbol    = byte('(')
	endArgsSymbol      = byte(')')
	argSeparatorSymbol = byte(',')
)

var (
	quoteSymbols = [...]byte{byte('"'), byte('\''), byte('`')}
)

//Given index i points to a beginning of a quote, returns the associated string - and the rest
// of the string that was unparsed
func getString(s string) (string, string, error) {
	i := 1
	quotechar := s[0]
	for ; len(s) > i && s[i] != quotechar; i++ {

	}
	if i >= len(s) || s[i] != quotechar {
		return "", s, errors.New("Could not find end of string - end quote is missing.")
	}

	return s[1:i], s[i+1:], nil
}

//gets rid of space characters from the beginning of the string
func eatSpace(s string) string {
	i := 0
	for ; len(s) > i && (s[i] == byte(' ') || s[i] == byte(' ') || s[i] == byte('\t') || s[i] == byte('\n') || s[i] == byte('\r')); i++ {

	}
	return s[i:]
}

//getSymbol gets the next symbol from the start of the string
func getSymbol(s string) (symb string, rem string, err error) {
	s = eatSpace(s)

	i := 0
	if len(s) == 0 {
		return "", s, nil
	}
	for ; i < len(quoteSymbols); i++ {
		if s[0] == quoteSymbols[i] {
			return getString(s)
		}
	}
	if s[0] == pipeSymbol || s[0] == startArgsSymbol || s[0] == endArgsSymbol || s[0] == argSeparatorSymbol {
		return "", s, errors.New("Syntax error")
	}
	for i = 0; len(s) > i && !(s[i] == pipeSymbol || s[i] == startArgsSymbol || s[i] == endArgsSymbol || s[i] == argSeparatorSymbol); i++ {

	}
	return s[:i], s[i:], nil
}

//PipelineElement represents a function call with optional arguments
type PipelineElement struct {
	Symbol string
	Args   []string
}

//ParsePipeline gets a pipeline string, and
func ParsePipeline(s string) (pipeline []PipelineElement, err error) {
	pipeline = make([]PipelineElement, 0)
	for s != "" {
		var pe PipelineElement
		//Start with symbol
		pe.Symbol, s, err = getSymbol(s)
		if err != nil {
			return nil, err
		}

		if len(s) > 0 && s[0] == startArgsSymbol {
			pe.Args = make([]string, 0)
			symb := ""
			symb, s, err = getSymbol(s[1:])
			if err != nil {
				return nil, err
			}
			pe.Args = append(pe.Args, symb)
			s = eatSpace(s)

			for len(s) > 0 && s[0] == argSeparatorSymbol {
				symb, s, err = getSymbol(s[1:])
				if err != nil {
					return nil, err
				}
				pe.Args = append(pe.Args, symb)
				s = eatSpace(s)
			}

			if !(len(s) > 0 && s[0] == endArgsSymbol) {
				return nil, errors.New("End ')' not found")
			}
			s = s[1:]
		}
		if len(s) > 1 {
			s = s[1:]
		}
		pipeline = append(pipeline, pe)
	}
	return pipeline, nil
}
