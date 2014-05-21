package generator

import (
	"bytes"
	"go/ast"
	"go/printer"
	"go/token"
)

func GenerateFake(structName, packageName string, interfaceNode *ast.InterfaceType) (string, error) {
	gen := generator{
		structName:    structName,
		packageName:   packageName,
		interfaceNode: interfaceNode,
	}

	buf := new(bytes.Buffer)
	err := printer.Fprint(buf, token.NewFileSet(), gen.File())
	return buf.String(), err
}

type generator struct {
	structName    string
	packageName   string
	interfaceNode *ast.InterfaceType
}

func (gen *generator) File() ast.Node {
	return &ast.File{
		Name: &ast.Ident{Name: gen.packageName},
		Decls: append([]ast.Decl{
			gen.typeDecl(),
			gen.constructorDecl(),
		}, gen.methodDecls()...),
	}
}

func (gen *generator) typeDecl() ast.Decl {
	return &ast.GenDecl{
		Tok: token.TYPE,
		Specs: []ast.Spec{
			&ast.TypeSpec{
				Name: &ast.Ident{Name: gen.structName},
				Type: &ast.StructType{
					Fields: &ast.FieldList{
						List: gen.structFields(),
					},
				},
			},
		},
	}
}

func (gen *generator) constructorDecl() ast.Decl {
	name := ast.NewIdent("New" + gen.structName)
	return &ast.FuncDecl{
		Name: name,
		Type: &ast.FuncType{
			Results: &ast.FieldList{
				List: []*ast.Field{
					{
						Type: &ast.StarExpr{X: ast.NewIdent(gen.structName)},
					},
				},
			},
		},
		Body: &ast.BlockStmt{
			List: []ast.Stmt{
				&ast.ReturnStmt{
					Results: []ast.Expr{
						&ast.UnaryExpr{
							Op: token.AND,
							X: &ast.CompositeLit{
								Type: ast.NewIdent(gen.structName),
								Elts: []ast.Expr{},
							},
						},
					},
				},
			},
		},
	}
}

func (gen *generator) methodDecls() []ast.Decl {
	result := []ast.Decl{}
	for _, method := range gen.interfaceNode.Methods.List {
		methodType := method.Type.(*ast.FuncType)
		paramNames := []string{}
		for _, field := range methodType.Params.List {
			if len(field.Names) > 0 {
				paramNames = append(paramNames, field.Names[0].Name)
			} else {
				panic("Don't handle anonymous args yet!")
			}
		}

		forwardArgs := []ast.Expr{}
		for _, name := range paramNames {
			forwardArgs = append(forwardArgs, ast.NewIdent(name))
		}

		forwardCall := &ast.CallExpr{
			Fun: &ast.SelectorExpr{
				X:   receiverIdent(),
				Sel: fakeFieldName(method.Names),
			},
			Args: forwardArgs,
		}

		var callStatement ast.Stmt
		if methodType.Results != nil {
			callStatement = &ast.ReturnStmt{
				Results: []ast.Expr{forwardCall},
			}
		} else {
			callStatement = &ast.ExprStmt{
				X: forwardCall,
			}
		}

		result = append(result, &ast.FuncDecl{
			Name: method.Names[0],
			Type: methodType,
			Recv: &ast.FieldList{
				List: []*ast.Field{
					{
						Names: []*ast.Ident{receiverIdent()},
						Type:  &ast.StarExpr{X: ast.NewIdent(gen.structName)},
					},
				},
			},
			Body: &ast.BlockStmt{
				List: []ast.Stmt{
					callStatement,
				},
			},
		})
	}
	return result
}

func (gen *generator) structFields() []*ast.Field {
	result := []*ast.Field{}
	for _, method := range gen.interfaceNode.Methods.List {
		name := fakeFieldName(method.Names)
		result = append(result, &ast.Field{
			Names: []*ast.Ident{name},
			Type:  method.Type,
		})
	}
	return result
}

func receiverIdent() *ast.Ident {
	return ast.NewIdent("fake")
}

func fakeFieldName(realNames []*ast.Ident) *ast.Ident {
	return ast.NewIdent(realNames[0].Name + "_")
}
