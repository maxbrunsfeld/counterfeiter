package locator

import (
	"errors"
	"go/ast"
	"go/parser"
	"go/token"
	"io/ioutil"
	"path"
	"strings"

	"github.com/maxbrunsfeld/counterfeiter/astutil"
	"github.com/maxbrunsfeld/counterfeiter/model"
)

func GetFunctionsFromDirectory(packageName, directory string) (*model.PackageToInterfacify, error) {
	funcs, err := GetFuncDecls(packageName, directory)
	if err != nil {
		panic(err)
	}

	return &model.PackageToInterfacify{
		Name:       packageName,
		ImportPath: directory,
		Funcs:      funcs,
	}, nil
}

func GetFuncDecls(packageName, directory string) ([]*ast.FuncDecl, error) {
	files, err := ioutil.ReadDir(directory)
	if err != nil {
		panic(err)
	}

	packages, err := packagesForDirPath(directory)
	if err != nil {
		panic(err)
	}

	types := getTypeNames(packages[packageName])

	fset := token.NewFileSet()
	funcSet := map[string]struct{}{}
	funcs := []*ast.FuncDecl{}
	for _, file := range files {

		if file.IsDir() || strings.HasSuffix(file.Name(), "_test.go") {
			continue
		}

		astFile, err := parser.ParseFile(fset, path.Join(directory, file.Name()), nil, parser.AllErrors)
		if err != nil {
			return nil, errors.New("failed to parse files in package: " + err.Error())
		}

		for _, decl := range astFile.Decls {
			switch fun := decl.(type) {
			case *ast.FuncDecl:
				// ignore functions with reciever types
				if fun.Recv != nil {
					continue
				}

				if !ast.IsExported(fun.Name.Name) {
					continue
				}

				// strip function body
				fun.Body = nil

				// unambiguate packages
				astutil.AddPackagePrefix(fun.Type, packageName, types)

				// just an ordinary function
				if _, ok := funcSet[fun.Name.Name]; !ok {
					funcSet[fun.Name.Name] = struct{}{}
					funcs = append(funcs, fun)
				}
			}
		}
	}

	return funcs, nil
}

func simpleImporter(imports map[string]*ast.Object, path string) (*ast.Object, error) {
	pkg := imports[path]
	if pkg == nil {
		// Guess the package name without importing it. Start with the last
		// element of the path.
		name := path[strings.LastIndex(path, "/")+1:]
		// Trim commonly used prefixes and suffixes containing illegal name
		// runes.
		name = strings.TrimSuffix(name, ".go")
		name = strings.TrimSuffix(name, "-go")
		name = strings.TrimPrefix(name, "go.")
		name = strings.TrimPrefix(name, "go-")
		name = strings.TrimPrefix(name, "biogo.")
		// It's also common for the last element of the path to contain an
		// extra "go" prefix, but not always. TODO: examine unresolved ids to
		// detect when trimming the "go" prefix is appropriate.
		pkg = ast.NewObj(ast.Pkg, name)
		pkg.Data = ast.NewScope(nil)
		imports[path] = pkg
	}
	return pkg, nil
}
