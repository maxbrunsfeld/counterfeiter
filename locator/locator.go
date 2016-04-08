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
	"unicode"

	"github.com/maxbrunsfeld/counterfeiter/model"
)

type InterfaceLocator interface {
	GetInterfacesFromFilePath(string) []string
}

func NewInterfaceLocator() InterfaceLocator {
	return interfaceLocator{}
}

type interfaceLocator struct{}

func (locator interfaceLocator) GetInterfacesFromFilePath(path string) []string {
	dir, err := getDir(path)
	if err != nil {
		panic(err)
	}

	importPath, err := importPathForDirPath(dir)
	if err != nil {
		panic(err)
	}

	dirPath, err := dirPathForImportPath(importPath)
	if err != nil {
		panic(err)
	}

	packages, err := packagesForDirPath(dirPath)
	if err != nil {
		panic(err)
	}

	interfacesInPackage := []string{}
	for _, pkg := range packages {

		for _, f := range pkg.Files {
			ast.Inspect(f, func(node ast.Node) bool {
				if typeSpec, ok := node.(*ast.TypeSpec); ok {
					if _, ok := typeSpec.Type.(*ast.InterfaceType); ok {
						firstRune := rune(typeSpec.Name.Name[0])

						if !unicode.IsUpper(firstRune) {
							return true
						}

						interfacesInPackage = append(interfacesInPackage, typeSpec.Name.Name)
					}
				}

				return true
			})
		}
	}

	return interfacesInPackage
}

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
		iface, file, isFunction, err := findInterface(pkg, interfaceName)
		if err != nil {
			return nil, err
		}

		typeNames := getTypeNames(pkg)

		if iface != nil {
			var methods []*ast.Field
			var err error
			var imports []*ast.ImportSpec = getImports(file)

			switch iface.(type) {
			case *ast.InterfaceType:
				interfaceNode := iface.(*ast.InterfaceType)
				methods, err = methodsForInterface(interfaceNode, importPath, pkg.Name, imports, typeNames)
			case *ast.FuncType:
				funcNode := iface.(*ast.FuncType)
				methods, err = methodsForFunction(funcNode, interfaceName, pkg.Name, typeNames)
			default:
				err = fmt.Errorf("cannot generate a counterfeit for a '%T'", iface)
			}

			if err != nil {
				return nil, err
			}

			return &model.InterfaceToFake{
				Name:                   interfaceName,
				Methods:                methods,
				ImportPath:             importPath,
				ImportSpecs:            imports,
				PackageName:            pkg.Name,
				RepresentedByInterface: !isFunction,
			}, nil
		}
	}

	return nil, fmt.Errorf("Could not find interface '%s'", interfaceName)
}

func addPackagePrefixToTypesInGeneratedPackage(t *ast.FuncType, pkgName string, typeNames map[string]bool) {
	ast.Inspect(t, func(node ast.Node) bool {
		switch node := node.(type) {
		case *ast.Field:
			addPackagePrefixToNode(&node.Type, pkgName, typeNames)
		case *ast.StarExpr:
			addPackagePrefixToNode(&node.X, pkgName, typeNames)
		case *ast.MapType:
			addPackagePrefixToNode(&node.Key, pkgName, typeNames)
			addPackagePrefixToNode(&node.Value, pkgName, typeNames)
		case *ast.ArrayType:
			addPackagePrefixToNode(&node.Elt, pkgName, typeNames)
		case *ast.ChanType:
			addPackagePrefixToNode(&node.Value, pkgName, typeNames)
		case *ast.Ellipsis:
			addPackagePrefixToNode(&node.Elt, pkgName, typeNames)
		}
		return true
	})
}

func addPackagePrefixToNode(node *ast.Expr, pkgName string, typeNames map[string]bool) {
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
	result = append(result, filepath.Join(runtime.GOROOT(), "src"))
	return result
}

func packagesForDirPath(path string) (map[string]*ast.Package, error) {
	return parser.ParseDir(token.NewFileSet(), path, nil, parser.AllErrors)
}

func findInterface(pkg *ast.Package, interfaceName string) (ast.Node, *ast.File, bool, error) {
	var file *ast.File
	var iface ast.Node
	var err error
	var isFunction bool

	for _, f := range pkg.Files {
		ast.Inspect(f, func(node ast.Node) bool {
			typeSpec, ok := node.(*ast.TypeSpec)
			if !ok || typeSpec.Name.Name != interfaceName {
				return true
			}

			switch typeSpec.Type.(type) {
			case *ast.InterfaceType:
				file = f
				iface = typeSpec.Type
			case *ast.FuncType:
				file = f
				isFunction = true
				iface = typeSpec.Type
			default:
				err = fmt.Errorf("Name '%s' does not refer to an interface", interfaceName)
			}
			return false
		})

		if iface != nil {
			break
		}
	}

	return iface, file, isFunction, err
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

func getTypeNames(pkg *ast.Package) map[string]bool {
	result := make(map[string]bool)
	ast.Inspect(pkg, func(node ast.Node) bool {
		if typeSpec, ok := node.(*ast.TypeSpec); ok {
			if typeSpec.Name != nil {
				result[typeSpec.Name.Name] = true
			}
		}
		return true
	})
	return result
}
