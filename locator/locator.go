package locator

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/maxbrunsfeld/counterfeiter/model"
)

func GetInterfaceFromFilePath(interfaceName, filePath string) (*model.InterfaceToFake, error) {
	dirPath, err := getDir(filePath)
	if err != nil {
		return nil, err
	}

	importPath, err := importPathForDirPath(dirPath)
	if err != nil {
		return nil, err
	}

	return getInterfaceFromImportPath(interfaceName, importPath)
}

func getInterfaceFromImportPath(interfaceName, importPath string) (*model.InterfaceToFake, error) {
	dirPath, err := dirPathForImportPath(importPath)
	if err != nil {
		return nil, err
	}

	packages, err := packagesForDirPath(dirPath)
	if err != nil {
		return nil, err
	}

	for _, pkg := range packages {
		iface, file, err := findInterface(pkg, interfaceName)
		if err != nil {
			return nil, err
		}

		typeNames := getTypeNames(pkg)

		if iface != nil {
			imports := getImports(file)
			methods, err := methodsForInterface(iface, importPath, pkg.Name, imports, typeNames)
			if err != nil {
				return nil, err
			}

			return &model.InterfaceToFake{
				Name:        interfaceName,
				Methods:     methods,
				ImportPath:  importPath,
				ImportSpecs: imports,
				PackageName: pkg.Name,
			}, nil
		}
	}

	return nil, fmt.Errorf("Could not find interface '%s'", interfaceName)
}

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

func prefixTypes(t *ast.FuncType, pkgName string, typeNames map[string]struct{}) {
	ast.Inspect(t, func(node ast.Node) bool {
		switch node := node.(type) {
		case *ast.Field:
			prefixType(&node.Type, pkgName, typeNames)
		case *ast.StarExpr:
			prefixType(&node.X, pkgName, typeNames)
		case *ast.MapType:
			prefixType(&node.Key, pkgName, typeNames)
			prefixType(&node.Value, pkgName, typeNames)
		case *ast.ArrayType:
			prefixType(&node.Elt, pkgName, typeNames)
		case *ast.ChanType:
			prefixType(&node.Value, pkgName, typeNames)
		}
		return true
	})
}

func prefixType(node *ast.Expr, pkgName string, typeNames map[string]struct{}) {
	if typeIdent, ok := (*node).(*ast.Ident); ok {
		if _, ok := typeNames[typeIdent.Name]; ok {
			*node = &ast.SelectorExpr{
				X:   ast.NewIdent(pkgName),
				Sel: typeIdent,
			}
		}
	}
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

func findImportPath(importSpecs []*ast.ImportSpec, alias string) string {
	for _, spec := range importSpecs {
		importPath := strings.Trim(spec.Path.Value, `"`)
		if path.Base(importPath) == alias {
			return importPath
		}
	}
	return ""
}

func dirPathForImportPath(importPath string) (string, error) {
	for _, goSrcPath := range goSourcePaths() {
		dirPath := filepath.Join(goSrcPath, filepath.Clean(importPath))
		stat, err := os.Stat(dirPath)
		if err == nil && stat.IsDir() {
			return dirPath, nil
		}
	}

	return "", fmt.Errorf("Package '%s' not found on GOPATH", importPath)
}

func importPathForDirPath(sourcePath string) (string, error) {
	sourcePath, err := filepath.Abs(sourcePath)
	if err != nil {
		return "", err
	}

	for _, goSrcPath := range goSourcePaths() {
		if strings.HasPrefix(sourcePath, goSrcPath) {
			return sourcePath[len(goSrcPath)+1:], nil
		}
	}

	return "", fmt.Errorf("Path '%s' is not on GOPATH", sourcePath)
}

func goSourcePaths() []string {
	result := []string{}
	for _, path := range strings.Split(os.Getenv("GOPATH"), ":") {
		result = append(result, filepath.Join(path, "src"))
	}
	result = append(result, filepath.Join(runtime.GOROOT(), "src", "pkg"))
	return result
}

func packagesForDirPath(path string) (map[string]*ast.Package, error) {
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

func getTypeNames(pkg *ast.Package) map[string]struct{} {
	result := map[string]struct{}{}
	ast.Inspect(pkg, func(node ast.Node) bool {
		if typeSpec, ok := node.(*ast.TypeSpec); ok {
			if typeSpec.Name != nil {
				result[typeSpec.Name.Name] = struct{}{}
			}
		}
		return true
	})
	return result
}
