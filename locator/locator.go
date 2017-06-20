package locator

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
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

	return getInterfaceFromDirPath(interfaceName, importPath, dirPath, vendorPaths...)
}

func GetInterfaceFromImportPath(interfaceName, importPath string, vendorPaths ...string) (*model.InterfaceToFake, error) {
	dirPath, err := dirPathForImportPath(importPath, vendorPaths)
	if err != nil {
		return nil, err
	}

	// The vendor paths passed to this function are only used to find dirPath.
	// The new dirPath might have different vendorPaths.
	vendorPaths, err = vendorPathsForDirPath(dirPath)
	if err != nil {
		return nil, err
	}

	return getInterfaceFromDirPath(interfaceName, importPath, dirPath, vendorPaths...)
}

func getInterfaceFromDirPath(interfaceName, importPath, dirPath string, vendorPaths ...string) (*model.InterfaceToFake, error) {
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
			importSpecs, err := getImports(file, vendorPaths...)
			if err != nil {
				return nil, err
			}

			pkgImport := pkg.Name
			if strings.HasSuffix(importPath, pkg.Name) {
				pkgImport = "xyz123"
			}

			var methods []model.Method
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

func excludeTests(fi os.FileInfo) bool {
	return !strings.HasSuffix(fi.Name(), "_test.go")
}

func getImports(file *ast.File, vendorPaths ...string) (result map[string]*ast.ImportSpec, err error) {
	defer func() {
		if r := recover(); r != nil {
			if rerr, ok := r.(error); ok {
				err = rerr
			} else {
				err = fmt.Errorf("panic: %v", r)
			}
		}
	}()
	result = map[string]*ast.ImportSpec{}
	ast.Inspect(file, func(node ast.Node) bool {
		if importSpec, ok := node.(*ast.ImportSpec); ok {
			var alias string
			if importSpec.Name == nil {

				importPath, err := strconv.Unquote(importSpec.Path.Value)
				if err != nil {
					panic(fmt.Errorf("could not unquote %v: %v", importSpec.Path.Value, err))
				}
				dirPath, err := dirPathForImportPath(importPath, vendorPaths)
				if err != nil {
					panic(fmt.Errorf("could not get directory for import path %v: %v", importPath, err))
				}
				importedPackages, err := parser.ParseDir(token.NewFileSet(), dirPath, excludeTests, parser.PackageClauseOnly)
				if err != nil {
					panic(fmt.Errorf("could not parse directory %v: %v", dirPath, err))
				}
				pkgNames := make([]string, 0, len(importedPackages))
				for key := range importedPackages {
					if key == "main" {
						// Ignore any non-importable packages.
						// net/http has a package main that is excluded
						// using build constraints, but we can't check
						// build constraints here.
						// This works around that.
						continue
					}
					pkgNames = append(pkgNames, key)
				}
				if len(pkgNames) == 0 {
					panic(fmt.Errorf("No package found in %v", importPath))
				}
				if len(pkgNames) != 1 {
					panic(fmt.Errorf("Multiple packages found in %v: %v", importPath, pkgNames))
				}
				alias = pkgNames[0]
			} else {
				alias = importSpec.Name.Name
			}
			result[alias] = importSpec
		}
		return true
	})
	return
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
