package model

import "go/ast"

type Method struct {
	Imports map[string]*ast.ImportSpec
	Field   *ast.Field
}

type InterfaceToFake struct {
	Name                   string
	Methods                []Method
	ImportPath             string
	PackageName            string
	RepresentedByInterface bool
}
