package locator

import (
	"go/ast"
)

func methodsForFunction(funcName string, funcNode *ast.FuncType) ([]*ast.Field, error) {
	return []*ast.Field{{
		Names: []*ast.Ident{{Name: funcName}},
		Type:  funcNode,
	}}, nil
}
