package locator

import (
	"go/ast"
	"strings"

	"github.com/maxbrunsfeld/counterfeiter/astutil"
	"github.com/maxbrunsfeld/counterfeiter/model"
)

func methodsForInterface(
	iface *ast.InterfaceType,
	importPath,
	pkgName string,
	importSpecs map[string]*ast.ImportSpec,
	knownTypes map[string]bool,
) ([]model.Method, error) {
	result := []model.Method{}
	for _, field := range iface.Methods.List {
		switch t := field.Type.(type) {
		case *ast.FuncType:

			// this  will mutate the actual ast node to generate "correct code"
			// it ensures func signatures have the correct package name for
			// types that belong to the package we are generating code from
			// e.g.: change "Param" to "foo.Param" when Param belongs to pkg "foo"
			astutil.AddPackagePrefix(t, pkgName, knownTypes)
			result = append(result,
				model.Method{
					Imports: importSpecs,
					Field:   field,
				})

		case *ast.Ident:
			iface, err := GetInterfaceFromImportPath(t.Name, importPath)
			if err != nil {
				return nil, err
			}
			result = append(result, iface.Methods...)
		case *ast.SelectorExpr:
			pkgAlias := t.X.(*ast.Ident).Name

			var pkgImportPath string
			if importSpec, ok := importSpecs[pkgAlias]; ok {
				pkgImportPath = strings.Trim(importSpec.Path.Value, `"`)
			}

			iface, err := GetInterfaceFromImportPath(t.Sel.Name, pkgImportPath)
			if err != nil {
				return nil, err
			}
			result = append(result, iface.Methods...)
		}
	}
	return result, nil
}
