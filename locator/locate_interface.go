package locator

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"log"

	"github.com/maxbrunsfeld/counterfeiter/astutil"
	"github.com/maxbrunsfeld/counterfeiter/model"
	goast "golang.org/x/tools/go/ast/astutil"
	"golang.org/x/tools/go/packages"
)

func methodsForInterface(
	iface *ast.InterfaceType,
	importPath string,
	pkgName string,
	pkg *packages.Package,
	importSpecs map[string]*ast.ImportSpec,
	knownTypes map[string]bool,
) ([]model.Method, error) {
	log.Printf("\n\nMethods for %s\n", importPath)
	result := []model.Method{}
	for i, field := range iface.Methods.List {
		log.Printf("\n\nField %v:\n", i)
		switch t := field.Type.(type) {
		case *ast.FuncType:
			log.Println("FUNC_TYPE")
			// this  will mutate the actual ast node to generate "correct code"
			// it ensures func signatures have the correct package name for
			// types that belong to the package we are generating code from
			// e.g.: change "Param" to "foo.Param" when Param belongs to pkg "foo"
			log.Printf("adding package prefix [%s] to [%s]\n", pkgName, field.Names[0])
			astutil.AddPackagePrefix(t, pkgName, knownTypes)
			result = append(result,
				model.Method{
					Imports: importSpecs,
					Field:   field,
				})

		case *ast.Ident:
			log.Println("IDENT")
			iface, err := GetInterfaceFromImportPath(t.Name, importPath)
			if err != nil {
				return nil, err
			}
			result = append(result, iface.Methods...)
		case *ast.SelectorExpr:
			log.Println("SELECTOR_EXPR")
			pkgAlias := t.X.(*ast.Ident).Name
			log.Printf(">>>>>>>>> Package Alias: %s\n", pkgAlias)
			pkgImportPath := findImportPath(importSpecs, pkgAlias)
			pkgAtPath := pkg.Imports[pkgImportPath]
			if pkgAtPath == nil {
				return nil, fmt.Errorf("cannot find package with import path: %s", pkgImportPath)
			}
			iface, err := GetInterfaceFromImportPath(t.Sel.Name, pkgAtPath.PkgPath)
			if err != nil {
				return nil, err
			}

			result = append(result, iface.Methods...)
		}
	}
	return result, nil
}

// pathEnclosingInterval returns the types.Info of the package and ast.Node that
// contain source interval [start, end), and all the node's ancestors
// up to the AST root.  It searches the ast.Files of initPkg and the packages it imports.
//
// Modified from golang.org/x/tools/go/loader.
func pathEnclosingInterval(initPkg *packages.Package, start, end token.Pos) (*types.Info, []ast.Node) {
	pkgs := []*packages.Package{initPkg}
	for _, pkg := range initPkg.Imports {
		pkgs = append(pkgs, pkg)
	}
	for _, pkg := range pkgs {
		for _, f := range pkg.Syntax {
			if f.Pos() == token.NoPos {
				// This can happen if the parser saw
				// too many errors and bailed out.
				// (Use parser.AllErrors to prevent that.)
				continue
			}
			if !tokenFileContainsPos(pkg.Fset.File(f.Pos()), start) {
				continue
			}
			if path, _ := goast.PathEnclosingInterval(f, start, end); path != nil {
				return pkg.TypesInfo, path
			}
		}
	}
	return nil, nil
}

func tokenFileContainsPos(f *token.File, pos token.Pos) bool {
	p := int(pos)
	base := f.Base()
	return base <= p && p < base+f.Size()
}
