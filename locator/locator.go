package locator

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path"
	"path/filepath"
	"strings"
)

func GetInterface(interfaceName, path string) (string, *ast.InterfaceType, []*ast.ImportSpec, error) {
	path, err := getDir(path)
	if err != nil {
		return "", nil, nil, err
	}

	importPath, err := getImportPath(path)
	if err != nil {
		return "", nil, nil, err
	}

	packages, err := getPackages(path)
	if err != nil {
		return "", nil, nil, err
	}

	for _, pkg := range packages {
		iface, file, err := findInterface(pkg, interfaceName)
		if err != nil {
			return "", nil, nil, err
		}

		if iface != nil {
			return importPath, iface, getImports(file), nil
		}
	}

	return "", nil, nil, fmt.Errorf("Could not find interface '%s'", interfaceName)
}

func getDir(path string) (string, error) {
	stat, err := os.Stat(path)
	if err != nil {
		return "", err
	}

	if !stat.IsDir() {
		path = filepath.Dir(path)
	}

	return path, nil
}

func getImportPath(sourcePath string) (string, error) {
	sourcePath, err := filepath.Abs(sourcePath)
	if err != nil {
		return "", err
	}

	gopaths := strings.Split(os.Getenv("GOPATH"), ":")
	for _, gopath := range gopaths {
		gopath = filepath.ToSlash(gopath)
		srcPath := path.Join(gopath, "src")
		if strings.HasPrefix(sourcePath, srcPath) {
			return sourcePath[len(srcPath)+1:], nil
		}
	}

	return "", fmt.Errorf("Path '%s' is not on GOPATH", sourcePath)
}

func getPackages(path string) (map[string]*ast.Package, error) {
	return parser.ParseDir(token.NewFileSet(), path, nil, parser.AllErrors)
}

func findInterface(pkg *ast.Package, interfaceName string) (*ast.InterfaceType, *ast.File, error) {
	var file *ast.File
	var iface *ast.InterfaceType
	var err error

	for _, f := range pkg.Files {
		ast.Inspect(f, func(node ast.Node) bool {
			typeSpec, ok := node.(*ast.TypeSpec)
			if ok && typeSpec.Name.Name == interfaceName {
				if interfaceType, ok := typeSpec.Type.(*ast.InterfaceType); ok {
					iface = interfaceType
					file = f
				} else {
					err = fmt.Errorf("Name '%s' does not refer to an interface", interfaceName)
				}
				return false
			}
			return true
		})

		if iface != nil {
			break
		}
	}

	return iface, file, err
}

func getImports(file *ast.File) []*ast.ImportSpec {
	result := []*ast.ImportSpec{}
	ast.Inspect(file, func(node ast.Node) bool {
		if importSpec, ok := node.(*ast.ImportSpec); ok {
			result = append(result, importSpec)
		}
		return true
	})
	return result
}
