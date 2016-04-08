package locator

import (
	"go/ast"
)

func methodsForInterface(
	iface *ast.InterfaceType,
	importPath,
	pkgName string,
	importSpecs []*ast.ImportSpec,
	typenamesNeedingPackageAlias map[string]bool,
) ([]*ast.Field, error) {
	result := []*ast.Field{}
	for _, field := range iface.Methods.List {
		switch t := field.Type.(type) {
		case *ast.FuncType:

			// this  will mutate the actual ast node to generate "correct code"
			// it ensures func signatures have the correct package name for
			// types that belong to the package we are generating code from
			// e.g.: change "Param" to "foo.Param" when Param belongs to pkg "foo"
			addPackagePrefixToTypesInGeneratedPackage(t, pkgName, typenamesNeedingPackageAlias)
			result = append(result, field)

		case *ast.Ident:
			iface, err := getInterfaceFromImportPath(t.Name, importPath)
			if err != nil {
				return nil, err
			}
			result = append(result, iface.Methods...)
		case *ast.SelectorExpr:
			pkgAlias := t.X.(*ast.Ident).Name
			pkgImportPath := findImportPath(importSpecs, pkgAlias)
			iface, err := getInterfaceFromImportPath(t.Sel.Name, pkgImportPath)
			if err != nil {
				return nil, err
			}
			result = append(result, iface.Methods...)
		}
	}
	return result, nil
}
