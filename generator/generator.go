package generator

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/printer"
	"go/token"
	"regexp"
	"strings"
)

func GenerateFake(
	structName, packageName string,
	interfaceNode *ast.InterfaceType) (string, error) {
	gen := generator{
		structName:    structName,
		packageName:   packageName,
		interfaceNode: interfaceNode,
	}

	buf := new(bytes.Buffer)
	err := printer.Fprint(buf, token.NewFileSet(), gen.SourceFile())
	return prettifyCode(buf.String()), err
}

type generator struct {
	structName    string
	packageName   string
	interfaceNode *ast.InterfaceType
}

func (gen *generator) SourceFile() ast.Node {
	return &ast.File{
		Name: &ast.Ident{Name: gen.packageName},
		Decls: append([]ast.Decl{
			gen.imports(),
			gen.typeDecl(),
		}, gen.methodDecls()...),
	}
}

func (gen *generator) imports() ast.Decl {
	return &ast.GenDecl{
		Tok: token.IMPORT,
		Specs: []ast.Spec{&ast.ImportSpec{
			Path: &ast.BasicLit{
				Kind:  token.STRING,
				Value: `"sync"`,
			},
		}},
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

func (gen *generator) methodDecls() []ast.Decl {
	result := []ast.Decl{}
	for _, method := range gen.interfaceNode.Methods.List {
		methodType := method.Type.(*ast.FuncType)

		result = append(
			result,
			gen.methodImplementation(method),
			gen.callsListGetter(method),
		)

		if methodType.Results != nil {
			result = append(
				result,
				gen.returnsMethod(method),
			)
		}
	}
	return result
}

func (gen *generator) structFields() []*ast.Field {
	result := []*ast.Field{
		{
			Type: &ast.SelectorExpr{
				X:   ast.NewIdent("sync"),
				Sel: ast.NewIdent("RWMutex"),
			},
		},
	}

	for _, method := range gen.interfaceNode.Methods.List {
		methodType := method.Type.(*ast.FuncType)

		result = append(
			result,

			&ast.Field{
				Names: []*ast.Ident{methodImplFuncIdent(method)},
				Type:  method.Type,
			},

			&ast.Field{
				Names: []*ast.Ident{callsListFieldIdent(method)},
				Type: &ast.ArrayType{
					Elt: gen.argsStructTypeForMethod(methodType),
				},
			},
		)

		if methodType.Results != nil {
			result = append(result, &ast.Field{
				Names: []*ast.Ident{returnStructIdent(method)},
				Type:  gen.returnStructTypeForMethod(methodType),
			})
		}
	}

	return result
}

func (gen *generator) argsStructTypeForMethod(methodType *ast.FuncType) *ast.StructType {
	paramFields := []*ast.Field{}
	for i, field := range methodType.Params.List {
		paramFields = append(paramFields, &ast.Field{
			Type:  field.Type,
			Names: []*ast.Ident{ast.NewIdent(publicize(nameForMethodParam(i)))},
		})
	}

	return &ast.StructType{
		Fields: &ast.FieldList{List: paramFields},
	}
}

func (gen *generator) returnStructTypeForMethod(methodType *ast.FuncType) *ast.StructType {
	resultFields := []*ast.Field{}
	for i, field := range methodType.Results.List {
		resultFields = append(resultFields, &ast.Field{
			Type:  field.Type,
			Names: []*ast.Ident{ast.NewIdent(nameForMethodResult(i))},
		})
	}

	return &ast.StructType{
		Fields: &ast.FieldList{List: resultFields},
	}
}

func (gen *generator) methodImplementation(method *ast.Field) *ast.FuncDecl {
	methodType := method.Type.(*ast.FuncType)

	stubMethod := &ast.SelectorExpr{
		X:   receiverIdent(),
		Sel: methodImplFuncIdent(method),
	}

	forwardArgs := []ast.Expr{}
	methodParams := []*ast.Field{}
	for i, field := range methodType.Params.List {
		forwardArgs = append(forwardArgs, ast.NewIdent(nameForMethodParam(i)))
		methodParams = append(methodParams, &ast.Field{
			Names: []*ast.Ident{ast.NewIdent(nameForMethodParam(i))},
			Type:  field.Type,
		})
	}

	forwardCall := &ast.CallExpr{
		Fun:  stubMethod,
		Args: forwardArgs,
	}

	var callStatement ast.Stmt
	if methodType.Results != nil {
		returnFields := []ast.Expr{}
		for i, _ := range methodType.Results.List {
			returnFields = append(returnFields, &ast.SelectorExpr{
				X: &ast.SelectorExpr{
					X:   receiverIdent(),
					Sel: returnStructIdent(method),
				},
				Sel: ast.NewIdent(nameForMethodResult(i)),
			})
		}

		callStatement = &ast.IfStmt{
			Cond: &ast.BinaryExpr{
				X:  stubMethod,
				Op: token.NEQ,
				Y: &ast.BasicLit{
					Kind:  token.STRING,
					Value: "nil",
				},
			},
			Body: &ast.BlockStmt{
				List: []ast.Stmt{
					&ast.ReturnStmt{
						Results: []ast.Expr{forwardCall},
					},
				},
			},
			Else: &ast.BlockStmt{
				List: []ast.Stmt{
					&ast.ReturnStmt{
						Results: returnFields,
					},
				},
			},
		}
	} else {
		callStatement = &ast.IfStmt{
			Cond: &ast.BinaryExpr{
				X:  stubMethod,
				Op: token.NEQ,
				Y: &ast.BasicLit{
					Kind:  token.STRING,
					Value: "nil",
				},
			},
			Body: &ast.BlockStmt{
				List: []ast.Stmt{
					&ast.ExprStmt{
						X: forwardCall,
					},
				},
			},
		}
	}

	return &ast.FuncDecl{
		Name: method.Names[0],
		Type: &ast.FuncType{
			Params:  &ast.FieldList{List: methodParams},
			Results: methodType.Results,
		},
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
				&ast.ExprStmt{
					X: &ast.CallExpr{
						Fun: &ast.SelectorExpr{
							X:   receiverIdent(),
							Sel: ast.NewIdent("Lock"),
						},
					},
				},
				&ast.DeferStmt{
					Call: &ast.CallExpr{
						Fun: &ast.SelectorExpr{
							X:   receiverIdent(),
							Sel: ast.NewIdent("Unlock"),
						},
					},
				},
				&ast.AssignStmt{
					Tok: token.ASSIGN,
					Lhs: []ast.Expr{&ast.SelectorExpr{
						X:   receiverIdent(),
						Sel: callsListFieldIdent(method),
					}},
					Rhs: []ast.Expr{&ast.CallExpr{
						Fun: ast.NewIdent("append"),
						Args: []ast.Expr{
							&ast.SelectorExpr{
								X:   receiverIdent(),
								Sel: callsListFieldIdent(method),
							},
							&ast.CompositeLit{
								Type: gen.argsStructTypeForMethod(methodType),
								Elts: forwardArgs,
							},
						},
					}},
				},
				callStatement,
			},
		},
	}
}

func (gen *generator) callsListGetter(method *ast.Field) *ast.FuncDecl {
	return &ast.FuncDecl{
		Name: callsListMethodIdent(method),
		Type: &ast.FuncType{
			Results: &ast.FieldList{List: []*ast.Field{
				&ast.Field{
					Type: &ast.ArrayType{
						Elt: gen.argsStructTypeForMethod(method.Type.(*ast.FuncType)),
					},
				},
			}},
		},
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
				&ast.ExprStmt{
					X: &ast.CallExpr{
						Fun: &ast.SelectorExpr{
							X:   receiverIdent(),
							Sel: ast.NewIdent("RLock"),
						},
					},
				},
				&ast.DeferStmt{
					Call: &ast.CallExpr{
						Fun: &ast.SelectorExpr{
							X:   receiverIdent(),
							Sel: ast.NewIdent("RUnlock"),
						},
					},
				},
				&ast.ReturnStmt{
					Results: []ast.Expr{
						&ast.SelectorExpr{
							X:   receiverIdent(),
							Sel: callsListFieldIdent(method),
						},
					},
				},
			},
		},
	}
}

func (gen *generator) returnsMethod(method *ast.Field) *ast.FuncDecl {
	params := []*ast.Field{}
	structFields := []ast.Expr{}
	for i, result := range method.Type.(*ast.FuncType).Results.List {
		params = append(params, &ast.Field{
			Names: []*ast.Ident{ast.NewIdent(nameForMethodResult(i))},
			Type:  result.Type,
		})

		structFields = append(structFields, &ast.KeyValueExpr{
			Key:   ast.NewIdent(nameForMethodResult(i)),
			Value: ast.NewIdent(nameForMethodResult(i)),
		})
	}

	return &ast.FuncDecl{
		Name: returnMethodIdent(method),
		Type: &ast.FuncType{
			Params: &ast.FieldList{List: params},
		},
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
				&ast.AssignStmt{
					Tok: token.ASSIGN,
					Lhs: []ast.Expr{
						&ast.SelectorExpr{
							X:   receiverIdent(),
							Sel: returnStructIdent(method),
						},
					},
					Rhs: []ast.Expr{
						&ast.CompositeLit{
							Type: gen.returnStructTypeForMethod(method.Type.(*ast.FuncType)),
							Elts: structFields,
						},
					},
				},
			},
		},
	}
}

func nameForMethodResult(i int) string {
	return fmt.Sprintf("result%d", i+1)
}

func nameForMethodParam(i int) string {
	return fmt.Sprintf("arg%d", i+1)
}

func callsListMethodIdent(method *ast.Field) *ast.Ident {
	return ast.NewIdent(method.Names[0].Name + "Calls")
}

func callsListFieldIdent(method *ast.Field) *ast.Ident {
	return ast.NewIdent(privatize(callsListMethodIdent(method).Name))
}

func methodImplFuncIdent(method *ast.Field) *ast.Ident {
	return ast.NewIdent(method.Names[0].Name + "Stub")
}

func returnMethodIdent(method *ast.Field) *ast.Ident {
	return ast.NewIdent(method.Names[0].Name + "Returns")
}

func returnStructIdent(method *ast.Field) *ast.Ident {
	return ast.NewIdent(privatize(returnMethodIdent(method).Name))
}

func receiverIdent() *ast.Ident {
	return ast.NewIdent("fake")
}

func publicize(input string) string {
	return strings.ToUpper(input[0:1]) + input[1:]
}

func privatize(input string) string {
	return strings.ToLower(input[0:1]) + input[1:]
}

var funcRegexp = regexp.MustCompile("\n(func)")

func prettifyCode(code string) string {
	return funcRegexp.ReplaceAllString(code, "\n\n$1")
}
