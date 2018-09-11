package locator

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"log"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"go/build"

	"github.com/maxbrunsfeld/counterfeiter/model"
	"golang.org/x/tools/go/packages"
)

func GetInterfaceFromFilePath(interfaceName, filePath string) (*model.InterfaceToFake, error) {
	if !strings.HasSuffix(strings.ToLower(filePath), ".go") {
		return GetInterfaceFromImportPath(interfaceName, filePath)
	}
	cfg := &packages.Config{
		Mode: packages.LoadSyntax,
	}

	pkgs, err := packages.Load(cfg, fmt.Sprintf("contains:%s", filePath))
	if err != nil || len(pkgs) == 0 {
		return nil, fmt.Errorf("couldn't load package for file %q: %v", filePath, err)
	}
	return getInterfaceFromPackage(interfaceName, pkgs[0])
}

func getInterfaceFromPackage(interfaceName string, pkg *packages.Package) (*model.InterfaceToFake, error) {
	iface, file, isFunction, err := findInterface(pkg, interfaceName)
	if err != nil {
		return nil, err
	}

	if iface != nil {
		typesFound := getTypeNames(pkg)
		importSpecs := getImports(file)

		pkgImport := pkg.Name
		pkgPath := pkg.PkgPath
		if i := strings.Index(pkgPath, "/vendor/"); i != -1 {
			pkgPath = pkgPath[i+len("/vendor/"):]
		}

		if strings.HasSuffix(pkgPath, pkg.Name) {
			log.Println("changing package", pkg.PkgPath, "to xyz123")
			pkgImport = "xyz123"
		}

		var methods []model.Method
		var err error
		switch iface.(type) {
		case *ast.InterfaceType:
			log.Println("getting methods for interface")
			methods, err = methodsForInterface(iface.(*ast.InterfaceType), pkgPath, pkg.Name, pkg, importSpecs, typesFound)
		case *ast.FuncType:
			log.Println("getting methods for func")
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
				Value: strconv.Quote(pkgPath),
			},
		}

		for k, v := range importSpecs {
			log.Printf("%s > %s > %s\n", k, v.Name.Name, v.Path.Value)
		}

		return &model.InterfaceToFake{
			Name:                   interfaceName,
			Methods:                methods,
			ImportPath:             pkgPath,
			PackageName:            pkg.Name,
			RepresentedByInterface: !isFunction,
		}, nil
	}

	return nil, fmt.Errorf("Could not find interface '%s'", interfaceName)
}

func GetInterfaceFromImportPath(interfaceName, importPath string) (*model.InterfaceToFake, error) {
	cfg := &packages.Config{
		Mode: packages.LoadSyntax,
	}

	pkgs, err := packages.Load(cfg, importPath)
	if err != nil {
		return nil, fmt.Errorf("couldn't load package %q: %v", importPath, err)
	}
	return getInterfaceFromPackage(interfaceName, pkgs[0])
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
		log.Println(importSpec.Name)
		log.Println(importSpec.Path.Value)
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
			return filepath.ToSlash(sourcePath[len(goSrcPath)+1:]), nil
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
	separator := string(os.PathListSeparator)
	for _, path := range strings.Split(build.Default.GOPATH, separator) {
		result = append(result, filepath.Join(path, "src"))
	}
	result = append(result, filepath.Join(runtime.GOROOT(), "src"))
	return result
}

func packagesForDirPath(path string) (map[string]*ast.Package, error) {
	return parser.ParseDir(token.NewFileSet(), path, nil, parser.AllErrors)
}

func findInterface(pkg *packages.Package, name string) (ast.Node, *ast.File, bool, error) {
	var (
		file       *ast.File
		iface      ast.Node
		err        error
		isFunction bool
	)
	ifaceObj := pkg.Types.Scope().Lookup(name)
	if ifaceObj == nil {
		return nil, nil, false, fmt.Errorf("interface with name %s not found in package %s", name, pkg.Name)
	}
	_, nodes := pathEnclosingInterval(pkg, ifaceObj.Pos(), ifaceObj.Pos())
	for _, node := range nodes {
		switch x := node.(type) {
		case *ast.TypeSpec:
			if iface == nil && x.Name.Name == name {
				switch x.Type.(type) {
				case *ast.InterfaceType:
					iface = x.Type
				case *ast.FuncType:
					iface = x.Type
					isFunction = true
				}
			}
		case *ast.File:
			if file == nil {
				file = x
			}
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

func getTypeNames(pkg *packages.Package) map[string]bool {
	result := make(map[string]bool)
	scope := pkg.Types.Scope()
	for _, name := range scope.Names() {
		obj := scope.Lookup(name)
		t, ok := obj.(*types.TypeName)
		if obj.Exported() && ok {
			result[t.Name()] = true
		}
	}
	return result
}
