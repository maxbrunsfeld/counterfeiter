package astutil

import "go/ast"

func InjectAlias(t *ast.FuncType, importedSpecs map[string]*ast.ImportSpec, aliases map[string]string) {
	ast.Inspect(t, func(node ast.Node) bool {
		switch node := node.(type) {
		case *ast.Field:
			convertAlias(&node.Type, importedSpecs, aliases)
		case *ast.StarExpr:
			convertAlias(&node.X, importedSpecs, aliases)
		case *ast.MapType:
			convertAlias(&node.Key, importedSpecs, aliases)
			convertAlias(&node.Value, importedSpecs, aliases)
		case *ast.ArrayType:
			convertAlias(&node.Elt, importedSpecs, aliases)
		case *ast.ChanType:
			convertAlias(&node.Value, importedSpecs, aliases)
		case *ast.Ellipsis:
			convertAlias(&node.Elt, importedSpecs, aliases)
		}
		return true
	})
}

func convertAlias(node *ast.Expr, importedSpecs map[string]*ast.ImportSpec, aliases map[string]string) {
	if typeSel, ok := (*node).(*ast.SelectorExpr); ok {
		prefixIdent := typeSel.X.(*ast.Ident)
		spec, found := importedSpecs[prefixIdent.Name]
		if !found {
			return
		}

		newAlias := aliases[spec.Path.Value]
		if newAlias == "." {
			*node = typeSel.Sel
		} else {
			prefixIdent.Name = newAlias
		}
	}
}

func AddPackagePrefix(t *ast.FuncType, pkgName string, knownTypes map[string]bool) {
	ast.Inspect(t, func(node ast.Node) bool {
		switch node := node.(type) {
		case *ast.Field:
			addPackagePrefixToNode(&node.Type, pkgName, knownTypes)
		case *ast.StarExpr:
			addPackagePrefixToNode(&node.X, pkgName, knownTypes)
		case *ast.MapType:
			addPackagePrefixToNode(&node.Key, pkgName, knownTypes)
			addPackagePrefixToNode(&node.Value, pkgName, knownTypes)
		case *ast.ArrayType:
			addPackagePrefixToNode(&node.Elt, pkgName, knownTypes)
		case *ast.ChanType:
			addPackagePrefixToNode(&node.Value, pkgName, knownTypes)
		case *ast.Ellipsis:
			addPackagePrefixToNode(&node.Elt, pkgName, knownTypes)
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
