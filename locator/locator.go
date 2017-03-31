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
	"strconv"
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

	vendorPaths, err := vendorPathsForDirPath(dirPath)
	if err != nil {
		return nil, err
	}

	return GetInterfaceFromImportPath(interfaceName, importPath, vendorPaths...)
}

func GetInterfaceFromImportPath(interfaceName, importPath string, vendorPaths ...string) (*model.InterfaceToFake, error) {
	dirPath, err := dirPathForImportPath(importPath, vendorPaths)
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

		if iface != nil {
			typesFound := getTypeNames(pkg)
			importSpecs := getImports(file)

			pkgImport := pkg.Name
			if strings.HasSuffix(importPath, pkg.Name) {
				pkgImport = "xyz123"
			}

			var methods []model.Method
			var err error
			switch iface.(type) {
			case *ast.InterfaceType:
				interfaceNode := iface.(*ast.InterfaceType)
				methods, err = methodsForInterface(interfaceNode, importPath, pkgImport, importSpecs, typesFound, vendorPaths)
			case *ast.FuncType:
				funcNode := iface.(*ast.FuncType)
				methods, err = methodsForFunction(funcNode, interfaceName, pkgImport, importSpecs, typesFound)
			default:
				err = fmt.Errorf("cannot generate a counterfeit for a '%T'", iface)
			}

			if err != nil {
				return nil, err
			}

			importSpecs[pkgImport] = &ast.ImportSpec{
				Name: &ast.Ident{Name: pkgImport},
				Path: &ast.BasicLit{
					Kind:  token.STRING,
					Value: strconv.Quote(importPath),
				},
			}

			return &model.InterfaceToFake{
				Name:                   interfaceName,
				Methods:                methods,
				ImportPath:             importPath,
				PackageName:            pkg.Name,
				RepresentedByInterface: !isFunction,
			}, nil
		}
	}

	return nil, fmt.Errorf("Could not find interface '%s'", interfaceName)
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

func findImportPath(importSpecs map[string]*ast.ImportSpec, alias string) string {
	if importSpec, ok := importSpecs[alias]; ok {
		return strings.Trim(importSpec.Path.Value, `"`)
	}
	return ""
}

func dirPathForImportPath(importPath string, vendorPaths []string) (string, error) {
	for _, srcPath := range append(vendorPaths, goSourcePaths()...) {
		dirPath := filepath.Join(srcPath, filepath.Clean(importPath))
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

func vendorPathsForDirPath(dirPath string) ([]string, error) {
	dirPath, err := filepath.Abs(dirPath)
	if err != nil {
		return nil, err
	}

	vendorPaths := []string{}
	for _, goSrcPath := range goSourcePaths() {
		for strings.HasPrefix(dirPath, goSrcPath) {
			vendorPath := filepath.Join(dirPath, "vendor")
			stat, err := os.Stat(vendorPath)
			if err == nil && stat.IsDir() {
				vendorPaths = append(vendorPaths, vendorPath)
			}
			dirPath = filepath.Dir(dirPath)
		}
	}

	return vendorPaths, nil
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

func getImports(file *ast.File) map[string]*ast.ImportSpec {
	result := map[string]*ast.ImportSpec{}
	ast.Inspect(file, func(node ast.Node) bool {
		if importSpec, ok := node.(*ast.ImportSpec); ok {
			var alias string
			if importSpec.Name == nil {
				alias = path.Base(strings.Trim(importSpec.Path.Value, `"`))
			} else {
				alias = importSpec.Name.Name
			}
			result[alias] = importSpec
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
