package locator

import (
	"go/ast"

	"github.com/maxbrunsfeld/counterfeiter/astutil"
	"github.com/maxbrunsfeld/counterfeiter/model"
)

func methodsForFunction(
	funcNode *ast.FuncType,
	funcName string,
	pkgName string,
	importSpecs map[string]*ast.ImportSpec,
	knownTypes map[string]bool,
) ([]model.Method, error) {

	// this  will mutate the actual ast node to generate "correct code"
	// it ensures func signatures have the correct package name for
	// types that belong to the package we are generating code from
	// e.g.: change "Param" to "foo.Param" when Param belongs to pkg "foo"
	astutil.AddPackagePrefix(funcNode, pkgName, knownTypes)

	return []model.Method{
		{
			Imports: importSpecs,
			Field: &ast.Field{
				Names: []*ast.Ident{{Name: funcName}},
				Type:  funcNode,
			},
		},
	}, nil
}
