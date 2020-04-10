package main

import (
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
)

func getStringVariable(filename string, varname string) string {
	// based on https://gist.github.com/ncdc/fef1099f54a655f8fb11daf86f7868b8
	// borrowed from https://github.com/lukehoban/go-outline/blob/master/main.go#L54-L107
	fset := token.NewFileSet()
	parserMode := parser.ParseComments
	var fileAst *ast.File
	var err error

	fileAst, err = parser.ParseFile(fset, filename, nil, parserMode)
	if err != nil {
		panic(err)
	}

	for _, d := range fileAst.Decls {
		switch decl := d.(type) {
		case *ast.GenDecl:
			for _, spec := range decl.Specs {
				switch spec := spec.(type) {
				case *ast.ValueSpec:
					for _, id := range spec.Names {
						if id.Name == varname {
							vstring := id.Obj.Decl.(*ast.ValueSpec).Values[0].(*ast.BasicLit).Value
							if vstring[0] == '"' || vstring[0] == '`' {
								return vstring[1 : len(vstring)-1]
							}
							return vstring
						}
					}
				default:

				}
			}
		}
	}
	return ""
}

func main() {
	flag.Parse()
	fmt.Print(getStringVariable(flag.Arg(0), flag.Arg(1)))
}
