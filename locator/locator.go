package locator

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
)

func GetInterface(interfaceName, path string) (*ast.InterfaceType, []*ast.ImportSpec, error) {
	stat, err := os.Stat(path)
	if err != nil {
		return nil, nil, err
	}
	if !stat.IsDir() {
		path = filepath.Dir(path)
	}

	packages, err := parser.ParseDir(token.NewFileSet(), path, nil, parser.AllErrors)
	if err != nil {
		return nil, nil, err
	}

	basename := filepath.Base(path)
	pkg := packages[basename]
	if pkg == nil {
		return nil, nil, fmt.Errorf("Couldn't find package '%s' in directory", basename)
	}

	importSpecs := []*ast.ImportSpec{}
	var result *ast.InterfaceType
	ast.Inspect(pkg, func(node ast.Node) bool {
		importSpec, ok := node.(*ast.ImportSpec)
		if ok {
			importSpecs = append(importSpecs, importSpec)
		}

		typeSpec, ok := node.(*ast.TypeSpec)
		if ok && typeSpec.Name.Name == interfaceName {
			if interfaceType, ok := typeSpec.Type.(*ast.InterfaceType); ok {
				result = interfaceType
			} else {
				err = fmt.Errorf("Name '%s' does not refer to an interface", interfaceName)
				return false
			}
		}
		return true
	})

	// usedImportSpecs := map[*ast.ImportSpec]struct{}{}
	// ast.Inspect(result, func(node ast.Node) bool {
	// if selector, ok := node.(*ast.SelectorExpr); ok {
	// fmt.Println("SELECTOR", selector)
	// }
	// return true
	// })

	if result == nil {
		return nil, nil, fmt.Errorf("Could not find interface '%s'", interfaceName)
	}

	return result, importSpecs, err
}
