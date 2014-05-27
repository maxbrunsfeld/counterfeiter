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
	structFields := []*ast.Field{
		{
			Type: &ast.SelectorExpr{
				X:   ast.NewIdent("sync"),
				Sel: ast.NewIdent("RWMutex"),
			},
		},
	}

	for _, method := range gen.interfaceNode.Methods.List {
		methodType := method.Type.(*ast.FuncType)

		structFields = append(
			structFields,

			&ast.Field{
				Names: []*ast.Ident{ast.NewIdent(methodStubFuncName(method))},
				Type:  method.Type,
			},

			&ast.Field{
				Names: []*ast.Ident{ast.NewIdent(callArgsFieldName(method))},
				Type: &ast.ArrayType{
					Elt: argsStructTypeForMethod(methodType),
				},
			},
		)

		if methodType.Results != nil {
			structFields = append(
				structFields,
				&ast.Field{
					Names: []*ast.Ident{ast.NewIdent(returnStructFieldName(method))},
					Type:  returnStructTypeForMethod(methodType),
				},
			)
		}
	}

	return &ast.GenDecl{
		Tok: token.TYPE,
		Specs: []ast.Spec{
			&ast.TypeSpec{
				Name: &ast.Ident{Name: gen.structName},
				Type: &ast.StructType{
					Fields: &ast.FieldList{
						List: structFields,
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
			gen.methodImplementationDecl(method),
			gen.methodCallCountGetterDecl(method),
			gen.methodCallArgsGetterDecl(method),
		)

		if methodType.Results != nil {
			result = append(
				result,
				gen.methodReturnsSetterDecl(method),
			)
		}
	}
	return result
}

func (gen *generator) methodImplementationDecl(method *ast.Field) *ast.FuncDecl {
	methodType := method.Type.(*ast.FuncType)

	stubMethod := &ast.SelectorExpr{
		X:   receiverIdent(),
		Sel: ast.NewIdent(methodStubFuncName(method)),
	}

	forwardArgs := []ast.Expr{}
	methodParams := []*ast.Field{}
	var ellipsisPos token.Pos

	for i, field := range methodType.Params.List {
		forwardArgs = append(forwardArgs, ast.NewIdent(nameForMethodParam(i)))

		if _, ok := field.Type.(*ast.Ellipsis); ok {
			ellipsisPos = token.Pos(i)
		}

		methodParams = append(methodParams, &ast.Field{
			Names: []*ast.Ident{ast.NewIdent(nameForMethodParam(i))},
			Type:  field.Type,
		})
	}

	forwardCall := &ast.CallExpr{
		Fun:      stubMethod,
		Args:     forwardArgs,
		Ellipsis: ellipsisPos,
	}

	var callStatement ast.Stmt
	if methodType.Results != nil {
		returnFields := []ast.Expr{}
		for i, _ := range methodType.Results.List {
			returnFields = append(returnFields, &ast.SelectorExpr{
				X: &ast.SelectorExpr{
					X:   receiverIdent(),
					Sel: ast.NewIdent(returnStructFieldName(method)),
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
						Sel: ast.NewIdent(callArgsFieldName(method)),
					}},
					Rhs: []ast.Expr{&ast.CallExpr{
						Fun: ast.NewIdent("append"),
						Args: []ast.Expr{
							&ast.SelectorExpr{
								X:   receiverIdent(),
								Sel: ast.NewIdent(callArgsFieldName(method)),
							},
							&ast.CompositeLit{
								Type: argsStructTypeForMethod(methodType),
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

func (gen *generator) methodCallCountGetterDecl(method *ast.Field) *ast.FuncDecl {
	return &ast.FuncDecl{
		Name: ast.NewIdent(callCountMethodName(method)),
		Type: &ast.FuncType{
			Results: &ast.FieldList{List: []*ast.Field{
				&ast.Field{
					Type: ast.NewIdent("int"),
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
						&ast.CallExpr{
							Fun: ast.NewIdent("len"),
							Args: []ast.Expr{
								&ast.SelectorExpr{
									X:   receiverIdent(),
									Sel: ast.NewIdent(callArgsFieldName(method)),
								},
							},
						},
					},
				},
			},
		},
	}
}

func (gen *generator) methodCallArgsGetterDecl(method *ast.Field) *ast.FuncDecl {
	indexIdent := ast.NewIdent("i")

	resultValues := []ast.Expr{}
	resultTypes := []*ast.Field{}

	for i, field := range method.Type.(*ast.FuncType).Params.List {
		resultValues = append(resultValues, &ast.SelectorExpr{
			X: &ast.IndexExpr{
				X: &ast.SelectorExpr{
					X:   receiverIdent(),
					Sel: ast.NewIdent(callArgsFieldName(method)),
				},
				Index: indexIdent,
			},
			Sel: ast.NewIdent(nameForMethodParam(i)),
		})

		resultTypes = append(resultTypes, &ast.Field{
			Type: storedTypeForType(field.Type),
		})
	}

	return &ast.FuncDecl{
		Name: ast.NewIdent(callArgsMethodName(method)),
		Type: &ast.FuncType{
			Params: &ast.FieldList{List: []*ast.Field{
				&ast.Field{
					Names: []*ast.Ident{indexIdent},
					Type:  ast.NewIdent("int"),
				},
			}},
			Results: &ast.FieldList{List: resultTypes},
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
					Results: resultValues,
				},
			},
		},
	}
}

func (gen *generator) methodReturnsSetterDecl(method *ast.Field) *ast.FuncDecl {
	params := []*ast.Field{}
	structFields := []ast.Expr{}
	for i, result := range method.Type.(*ast.FuncType).Results.List {
		params = append(params, &ast.Field{
			Names: []*ast.Ident{ast.NewIdent(nameForMethodResult(i))},
			Type:  result.Type,
		})

		structFields = append(structFields, ast.NewIdent(nameForMethodResult(i)))
	}

	return &ast.FuncDecl{
		Name: ast.NewIdent(returnSetterMethodName(method)),
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
							Sel: ast.NewIdent(returnStructFieldName(method)),
						},
					},
					Rhs: []ast.Expr{
						&ast.CompositeLit{
							Type: returnStructTypeForMethod(method.Type.(*ast.FuncType)),
							Elts: structFields,
						},
					},
				},
			},
		},
	}
}

func argsStructTypeForMethod(methodType *ast.FuncType) *ast.StructType {
	paramFields := []*ast.Field{}
	for i, field := range methodType.Params.List {
		paramFields = append(paramFields, &ast.Field{
			Type:  storedTypeForType(field.Type),
			Names: []*ast.Ident{ast.NewIdent(nameForMethodParam(i))},
		})
	}

	return &ast.StructType{
		Fields: &ast.FieldList{List: paramFields},
	}
}

func returnStructTypeForMethod(methodType *ast.FuncType) *ast.StructType {
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

func storedTypeForType(t ast.Expr) ast.Expr {
	if ellipsis, ok := t.(*ast.Ellipsis); ok {
		return &ast.ArrayType{Elt: ellipsis.Elt}
	} else {
		return t
	}
}

func nameForMethodResult(i int) string {
	return fmt.Sprintf("result%d", i+1)
}

func nameForMethodParam(i int) string {
	return fmt.Sprintf("arg%d", i+1)
}

func callCountMethodName(method *ast.Field) string {
	return method.Names[0].Name + "CallCount"
}

func callArgsMethodName(method *ast.Field) string {
	return method.Names[0].Name + "ArgsForCall"
}

func callArgsFieldName(method *ast.Field) string {
	return privatize(callArgsMethodName(method))
}

func methodStubFuncName(method *ast.Field) string {
	return method.Names[0].Name + "Stub"
}

func returnSetterMethodName(method *ast.Field) string {
	return method.Names[0].Name + "Returns"
}

func returnStructFieldName(method *ast.Field) string {
	return privatize(returnSetterMethodName(method))
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
	code = funcRegexp.ReplaceAllString(code, "\n\n$1")
	code = strings.Replace(code, "\n\n\n", "\n\n", -1)
	return code
}
