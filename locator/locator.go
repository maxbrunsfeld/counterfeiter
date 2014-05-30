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

	var file *ast.File
	var result *ast.InterfaceType
	for _, f := range pkg.Files {
		ast.Inspect(f, func(node ast.Node) bool {
			typeSpec, ok := node.(*ast.TypeSpec)
			if ok && typeSpec.Name.Name == interfaceName {
				if interfaceType, ok := typeSpec.Type.(*ast.InterfaceType); ok {
					result = interfaceType
					file = f
				} else {
					err = fmt.Errorf("Name '%s' does not refer to an interface", interfaceName)
				}
				return false
			}
			return true
		})
	}

	if result == nil {
		return nil, nil, fmt.Errorf("Could not find interface '%s'", interfaceName)
	}

	importSpecs := []*ast.ImportSpec{}
	ast.Inspect(file, func(node ast.Node) bool {
		importSpec, ok := node.(*ast.ImportSpec)
		if ok {
			importSpecs = append(importSpecs, importSpec)
		}
		return true
	})

	return result, importSpecs, err
}
