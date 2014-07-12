package model

import "go/ast"

type InterfaceToFake struct {
	Name        string
	Methods     []*ast.Field
	ImportSpecs []*ast.ImportSpec
	ImportPath  string
	PackageName string
}
