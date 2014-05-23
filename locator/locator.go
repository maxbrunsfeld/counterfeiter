package locator

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
)

func GetInterface(interfaceName, path string) (*ast.InterfaceType, error) {
	stat, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	if !stat.IsDir() {
		path = filepath.Dir(path)
	}

	packages, err := parser.ParseDir(token.NewFileSet(), path, nil, parser.AllErrors)
	if err != nil {
		return nil, err
	}

	basename := filepath.Base(path)
	pkg := packages[basename]
	if pkg == nil {
		return nil, fmt.Errorf("Couldn't find package '%s' in directory", basename)
	}

	var result *ast.InterfaceType
	ast.Inspect(pkg, func(node ast.Node) bool {
		if typeSpec, ok := node.(*ast.TypeSpec); ok {
			if typeSpec.Name.Name == interfaceName {
				if interfaceType, ok := typeSpec.Type.(*ast.InterfaceType); ok {
					result = interfaceType
				} else {
					err = fmt.Errorf("Name '%s' does not refer to an interface", interfaceName)
				}
				return false
			}
		}
		return true
	})

	if result == nil {
		return nil, fmt.Errorf("Could not find interface '%s'", interfaceName)
	}

	return result, err
}
