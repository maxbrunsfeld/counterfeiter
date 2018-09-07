package locator

import (
	"errors"
	"go/ast"
	"go/token"
	"go/types"
	"log"

	"github.com/maxbrunsfeld/counterfeiter/astutil"
	"github.com/maxbrunsfeld/counterfeiter/model"
	goast "golang.org/x/tools/go/ast/astutil"
	"golang.org/x/tools/go/packages"
	"golang.org/x/tools/go/types/typeutil"
)

func methodsForInterface2(
	iface types.Object,
	importPath string,
	pkg *packages.Package,
	importSpecs map[string]*ast.ImportSpec,
	knownTypes map[string]bool,
	vendorPaths []string,
) ([]model.Method, error) {
	result := []model.Method{}
	tn, ok := iface.(*types.TypeName)
	if !ok {
		return nil, errors.New(tn.Name() + " is not a TypeName")
	}

	methods := typeutil.IntuitiveMethodSet(iface.Type(), nil)
	for _, method := range methods {
		log.Println("processing method", method.Obj().Name())
		_, nodes := pathEnclosingInterval(pkg, method.Obj().Pos(), method.Obj().Pos())
		for _, node := range nodes {
			if field, ok := node.(*ast.Field); ok {
				result = append(result, model.Method{
					Field:   field,
					Imports: importSpecs,
				})
				break
			}
		}
	}
	return result, nil
}

func methodsForInterface(
	iface *ast.InterfaceType,
	importPath string,
	pkgName string,
	importSpecs map[string]*ast.ImportSpec,
	knownTypes map[string]bool,
) ([]model.Method, error) {
	result := []model.Method{}
	for _, field := range iface.Methods.List {
		switch t := field.Type.(type) {
		case *ast.FuncType:

			// this  will mutate the actual ast node to generate "correct code"
			// it ensures func signatures have the correct package name for
			// types that belong to the package we are generating code from
			// e.g.: change "Param" to "foo.Param" when Param belongs to pkg "foo"
			astutil.AddPackagePrefix(t, pkgName, knownTypes)
			result = append(result,
				model.Method{
					Imports: importSpecs,
					Field:   field,
				})

		case *ast.Ident:
			iface, err := GetInterfaceFromImportPath(t.Name, importPath)
			if err != nil {
				return nil, err
			}
			result = append(result, iface.Methods...)
		case *ast.SelectorExpr:
			pkgAlias := t.X.(*ast.Ident).Name
			pkgImportPath := findImportPath(importSpecs, pkgAlias)
			iface, err := GetInterfaceFromImportPath(t.Sel.Name, pkgImportPath)
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
