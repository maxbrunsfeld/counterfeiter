package locator

import (
	"go/ast"
)

func methodsForFunction(
	funcNode *ast.FuncType,
	funcName string,
	pkgName string,
	typenamesNeedingPackageAlias map[string]bool,
) ([]*ast.Field, error) {

	// this  will mutate the actual ast node to generate "correct code"
	// it ensures func signatures have the correct package name for
	// types that belong to the package we are generating code from
	// e.g.: change "Param" to "foo.Param" when Param belongs to pkg "foo"
	addPackagePrefixToTypesInGeneratedPackage(funcNode, pkgName, typenamesNeedingPackageAlias)

	return []*ast.Field{{
		Names: []*ast.Ident{{Name: funcName}},
		Type:  funcNode,
	}}, nil
}
