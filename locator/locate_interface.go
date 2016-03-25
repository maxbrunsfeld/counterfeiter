package locator

import (
	"go/ast"
)

func methodsForInterface(iface *ast.InterfaceType, importPath, pkgName string, importSpecs []*ast.ImportSpec, typeNames map[string]struct{}) ([]*ast.Field, error) {
	result := []*ast.Field{}
	for _, field := range iface.Methods.List {
		switch t := field.Type.(type) {
		case *ast.FuncType:
			prefixTypes(t, pkgName, typeNames)
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
