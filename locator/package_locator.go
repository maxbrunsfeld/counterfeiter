package locator

import (
	"errors"
	"go/ast"
	"go/parser"
	"io/ioutil"
	"path/filepath"
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

	packages, fset, err := packagesForDirPath(directory)
	if err != nil {
		panic(err)
	}

	types := getTypeNames(packages[packageName])

	funcSet := map[string]struct{}{}
	funcs := []*ast.FuncDecl{}

	for _, file := range files {
		if file.IsDir() || strings.HasSuffix(file.Name(), "_test.go") {
			continue
		}

		astFile, err := parser.ParseFile(fset, filepath.Join(directory, file.Name()), nil, parser.AllErrors)
		if err != nil {
			return nil, errors.New("failed to parse files in package: " + err.Error())
		}

		for _, decl := range astFile.Decls {
			switch fun := decl.(type) {
			case *ast.FuncDecl:
				// ignore functions with receiver types
				if fun.Recv != nil {
					continue
				}

				if !ast.IsExported(fun.Name.Name) {
					continue
				}

				// strip function body
				fun.Body = nil

				// disambiguate packages
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
