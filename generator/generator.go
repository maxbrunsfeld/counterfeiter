package generator

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/token"
	"regexp"
	"strconv"
	"strings"

	"github.com/maxbrunsfeld/counterfeiter/astutil"
	"github.com/maxbrunsfeld/counterfeiter/model"

	"golang.org/x/tools/imports"
)

type CodeGenerator struct {
	Model       model.InterfaceToFake
	StructName  string
	PackageName string

	packageAlias map[string]string
}

func (gen CodeGenerator) GenerateFake() (string, error) {
	buf := new(bytes.Buffer)
	err := format.Node(buf, token.NewFileSet(), gen.buildASTForFake())
	if err != nil {
		return "", err
	}

	code, err := imports.Process("", buf.Bytes(), nil)
	return commentLine() + prettifyCode(string(code)), err
}

func (gen CodeGenerator) isExportedInterface() bool {
	return ast.IsExported(gen.Model.Name)
}

// The anatomy of a generated fake
// FIXME: These would be good to break into separate builders
//        (they could be individually unit tested, and this would just delegate)
/*
  imports
  type MySpecialFake struct {}
  MyMethod()
  MyMethodCallCount()
  MyMethodArgsForCall()
  Invocations -> map[string][][]interfac{}
  recordInvocation(string, []interface{})
  var _ fixtures.SomeInterface = new(MySpecialFake)
*/

func (gen CodeGenerator) buildASTForFake() ast.Node {
	gen.packageAlias = map[string]string{}

	declarations := []ast.Decl{}
	declarations = append(declarations, gen.imports())
	gen.fixup()

	declarations = append(declarations, gen.fakeStructDeclaration())

	for _, m := range gen.Model.Methods {
		methodType := m.Field.Type.(*ast.FuncType)

		declarations = append(
			declarations,
			gen.stubbedMethodImplementation(m.Field),
			gen.methodCallCountGetter(m.Field),
		)

		if methodType.Params.NumFields() > 0 {
			declarations = append(
				declarations,
				gen.methodCallArgsGetter(m.Field),
			)
		}

		if methodType.Results.NumFields() > 0 {
			declarations = append(
				declarations,
				gen.methodReturnsSetter(m.Field),
				gen.methodReturnsOnCallSetter(m.Field),
			)
		}
	}

	declarations = append(declarations, gen.recordedInvocationsMethod())
	declarations = append(declarations, gen.recordInvocationMethod())

	if gen.isExportedInterface() {
		declarations = append(
			declarations,
			gen.ensureInterfaceIsUsed(),
		)
	}

	return &ast.File{
		Name:  &ast.Ident{Name: gen.PackageName},
		Decls: declarations,
	}
}

func (gen CodeGenerator) imports() ast.Decl {
	specs := []ast.Spec{}
	allImports := map[string]bool{}
	dotImports := map[string]bool{}
	aliasImportNames := map[string]string{}

	modelImportName := strconv.Quote(gen.Model.ImportPath)
	allImports[modelImportName] = true

	syncImportName := strconv.Quote("sync")
	allImports[syncImportName] = true
	gen.packageAlias[syncImportName] = "sync"

	for _, m := range gen.Model.Methods {
		for packageName, importSpec := range m.Imports {
			if packageName == "." {
				dotImports[importSpec.Name.Name] = true
				gen.packageAlias[importSpec.Path.Value] = "."
			}

			allImports[importSpec.Path.Value] = true

			var importAlias = ""
			if importSpec.Name != nil && importSpec.Name.Name != "xyz123" {
				importAlias = importSpec.Name.Name
			}
			aliasImportNames[importSpec.Path.Value] = importAlias
		}
	}

	aliases := map[string]bool{}
	aliases[gen.Model.PackageName] = true
	gen.packageAlias[modelImportName] = gen.Model.PackageName
	for importName := range allImports {
		if _, found := gen.packageAlias[importName]; found {
			continue
		}

		alias := aliasImportNames[importName]
		if alias == "" {
			alias = gen.generateAlias(importName, aliases)
			if alias == "" {
				panic("could not generate an alias for " + importName)
			}
		}
		aliases[alias] = true
		gen.packageAlias[importName] = alias
	}

	for importName, alias := range gen.packageAlias {
		var name *ast.Ident
		if !strings.HasSuffix(importName[:len(importName)-1], alias) {
			name = &ast.Ident{Name: alias}
		}
		specs = append(specs, &ast.ImportSpec{
			Name: name,
			Path: &ast.BasicLit{
				Kind:  token.STRING,
				Value: importName,
			},
		})
	}

	return &ast.GenDecl{
		Lparen: 1,
		Tok:    token.IMPORT,
		Specs:  specs,
	}
}

var identifierRegex = regexp.MustCompile(`[^[:alnum:]]`)

func (gen CodeGenerator) generateAlias(importName string, aliases map[string]bool) string {
	unquoted, err := strconv.Unquote(importName)
	if err != nil {
		panic("cannot generate alias for " + importName)
	}

	paths := strings.Split(unquoted, "/")
	alias := ""
	for i := len(paths) - 1; i >= 0; i-- {
		safePath := identifierRegex.ReplaceAllString(paths[i], "_")

		alias = alias + safePath

		if aliases[alias] == false {
			return alias
		}
	}

	return ""
}

func (gen CodeGenerator) fixup() {
	for _, m := range gen.Model.Methods {
		typ := m.Field.Type.(*ast.FuncType)
		astutil.InjectAlias(typ, m.Imports, gen.packageAlias)
	}
}

func (gen CodeGenerator) fakeStructDeclaration() ast.Decl {
	structFields := []*ast.Field{}

	for _, m := range gen.Model.Methods {
		methodType := m.Field.Type.(*ast.FuncType)

		structFields = append(
			structFields,

			&ast.Field{
				Names: []*ast.Ident{ast.NewIdent(gen.methodStubFuncName(m.Field))},
				Type:  m.Field.Type,
			},

			&ast.Field{
				Type: &ast.SelectorExpr{
					X:   ast.NewIdent("sync"),
					Sel: ast.NewIdent("RWMutex"),
				},
				Names: []*ast.Ident{ast.NewIdent(gen.mutexFieldName(m.Field))},
			},

			&ast.Field{
				Names: []*ast.Ident{ast.NewIdent(gen.callArgsFieldName(m.Field))},
				Type: &ast.ArrayType{
					Elt: argsStructTypeForMethod(methodType),
				},
			},
		)

		if methodType.Results.NumFields() > 0 {
			structFields = append(
				structFields,
				&ast.Field{
					Names: []*ast.Ident{ast.NewIdent(gen.returnStructFieldName(m.Field))},
					Type:  returnStructTypeForMethod(methodType),
				},
				&ast.Field{
					Names: []*ast.Ident{ast.NewIdent(gen.returnMapFieldName(m.Field))},
					Type:  returnMapTypeForMethod(methodType),
				},
			)
		}
	}

	// include list of invocations
	structFields = append(structFields, &ast.Field{
		Names: []*ast.Ident{ast.NewIdent("invocations")},
		Type: &ast.MapType{
			Key:   ast.NewIdent("string"),
			Value: ast.NewIdent("[][]interface{}"),
		},
	})
	// and mutex for recording invocations
	structFields = append(structFields, &ast.Field{
		Names: []*ast.Ident{ast.NewIdent("invocationsMutex")},
		Type: &ast.SelectorExpr{
			X:   ast.NewIdent("sync"),
			Sel: ast.NewIdent("RWMutex"),
		},
	})

	return &ast.GenDecl{
		Tok: token.TYPE,
		Specs: []ast.Spec{
			&ast.TypeSpec{
				Name: &ast.Ident{Name: gen.StructName},
				Type: &ast.StructType{
					Fields: &ast.FieldList{List: structFields},
				},
			},
		},
	}
}

func (gen CodeGenerator) stubbedMethodImplementation(method *ast.Field) *ast.FuncDecl {
	methodType := method.Type.(*ast.FuncType)

	stubFunc := &ast.SelectorExpr{
		X:   receiverIdent(),
		Sel: ast.NewIdent(gen.methodStubFuncName(method)),
	}

	paramValuesToRecord := []ast.Expr{}
	paramValuesToPassToStub := []ast.Expr{}
	paramFields := []*ast.Field{}
	var ellipsisPos token.Pos
	var bodyStatements []ast.Stmt

	eachMethodParam(methodType, func(name string, t ast.Expr, i int) {
		paramFields = append(paramFields, &ast.Field{
			Names: []*ast.Ident{ast.NewIdent(name)},
			Type:  t,
		})

		if _, ok := t.(*ast.Ellipsis); ok {
			ellipsisPos = token.Pos(i + 1)
		}

		if tArray, ok := t.(*ast.ArrayType); ok && tArray.Len == nil {
			copyName := name + "Copy"
			bodyStatements = append(bodyStatements,
				&ast.DeclStmt{
					Decl: &ast.GenDecl{
						Tok: token.VAR,
						Specs: []ast.Spec{
							&ast.ValueSpec{
								Names: []*ast.Ident{ast.NewIdent(copyName)},
								Type:  t,
							},
						},
					},
				},
				&ast.IfStmt{
					Cond: invertNilCheck(ast.NewIdent(name)),
					Body: &ast.BlockStmt{List: []ast.Stmt{
						&ast.AssignStmt{
							Tok: token.ASSIGN,
							Lhs: []ast.Expr{
								ast.NewIdent(copyName),
							},
							Rhs: []ast.Expr{
								&ast.CallExpr{
									Fun: ast.NewIdent("make"),
									Args: []ast.Expr{
										t,
										&ast.CallExpr{
											Fun:  ast.NewIdent("len"),
											Args: []ast.Expr{ast.NewIdent(name)},
										},
									},
								},
							},
						},
						&ast.ExprStmt{
							X: &ast.CallExpr{
								Fun: ast.NewIdent("copy"),
								Args: []ast.Expr{
									ast.NewIdent(copyName),
									ast.NewIdent(name),
								},
							},
						},
					}},
				})
			paramValuesToRecord = append(paramValuesToRecord, ast.NewIdent(copyName))
		} else {
			paramValuesToRecord = append(paramValuesToRecord, ast.NewIdent(name))
		}
		paramValuesToPassToStub = append(paramValuesToPassToStub, ast.NewIdent(name))
	})

	stubFuncCall := &ast.CallExpr{
		Fun:      stubFunc,
		Args:     paramValuesToPassToStub,
		Ellipsis: ellipsisPos,
	}

	var lastStatements []ast.Stmt
	if methodType.Results.NumFields() > 0 {
		returnValues := []ast.Expr{}
		specificReturnValues := []ast.Expr{}
		eachMethodResult(methodType, func(name string, t ast.Expr) {
			returnValues = append(returnValues, &ast.SelectorExpr{
				X: &ast.SelectorExpr{
					X:   receiverIdent(),
					Sel: ast.NewIdent(gen.returnStructFieldName(method)),
				},
				Sel: ast.NewIdent(name),
			})
			specificReturnValues = append(specificReturnValues, &ast.SelectorExpr{
				X:   ast.NewIdent("ret"),
				Sel: ast.NewIdent(name),
			})
		})

		lastStatements = []ast.Stmt{
			&ast.IfStmt{
				Cond: invertNilCheck(stubFunc),
				Body: &ast.BlockStmt{List: []ast.Stmt{
					&ast.ReturnStmt{Results: []ast.Expr{stubFuncCall}},
				}},
			},
			&ast.IfStmt{
				Cond: ast.NewIdent("specificReturn"),
				Body: &ast.BlockStmt{List: []ast.Stmt{
					&ast.ReturnStmt{Results: specificReturnValues},
				}},
			},
			&ast.ReturnStmt{Results: returnValues},
		}
	} else {
		lastStatements = []ast.Stmt{&ast.IfStmt{
			Cond: invertNilCheck(stubFunc),
			Body: &ast.BlockStmt{List: []ast.Stmt{
				&ast.ExprStmt{X: stubFuncCall},
			}},
		}}
	}

	bodyStatements = append(bodyStatements,
		gen.callMutex(method, "Lock"),
	)

	if methodType.Results.NumFields() > 0 {
		bodyStatements = append(bodyStatements,
			&ast.AssignStmt{
				Tok: token.DEFINE,
				Lhs: []ast.Expr{
					ast.NewIdent("ret"),
					ast.NewIdent("specificReturn"),
				},
				Rhs: []ast.Expr{
					&ast.IndexExpr{
						X: &ast.SelectorExpr{
							X:   receiverIdent(),
							Sel: ast.NewIdent(gen.returnMapFieldName(method)),
						},
						Index: &ast.CallExpr{
							Fun: ast.NewIdent("len"),
							Args: []ast.Expr{
								&ast.SelectorExpr{
									X:   receiverIdent(),
									Sel: ast.NewIdent(gen.callArgsFieldName(method)),
								},
							},
						},
					},
				},
			},
		)
	}

	bodyStatements = append(bodyStatements,
		&ast.AssignStmt{
			Tok: token.ASSIGN,
			Lhs: []ast.Expr{&ast.SelectorExpr{
				X:   receiverIdent(),
				Sel: ast.NewIdent(gen.callArgsFieldName(method)),
			}},
			Rhs: []ast.Expr{&ast.CallExpr{
				Fun: ast.NewIdent("append"),
				Args: []ast.Expr{
					&ast.SelectorExpr{
						X:   receiverIdent(),
						Sel: ast.NewIdent(gen.callArgsFieldName(method)),
					},
					&ast.CompositeLit{
						Type: argsStructTypeForMethod(methodType),
						Elts: paramValuesToRecord,
					},
				},
			}},
		},

		&ast.ExprStmt{
			X: &ast.CallExpr{
				Fun: &ast.SelectorExpr{
					X:   receiverIdent(),
					Sel: ast.NewIdent("recordInvocation"),
				},
				Args: []ast.Expr{quotedMethodName(method), &ast.CompositeLit{
					Type: ast.NewIdent("[]interface{}"),
					Elts: paramValuesToRecord,
				},
				},
			},
		},

		gen.callMutex(method, "Unlock"),
	)

	bodyStatements = append(bodyStatements, lastStatements...)

	var methodName *ast.Ident
	if gen.Model.RepresentedByInterface {
		methodName = method.Names[0]
	} else {
		methodName = ast.NewIdent("Spy")
	}

	return &ast.FuncDecl{
		Name: methodName,
		Type: &ast.FuncType{
			Params:  &ast.FieldList{List: paramFields},
			Results: methodType.Results,
		},
		Recv: gen.receiverFieldList(),
		Body: &ast.BlockStmt{List: bodyStatements},
	}
}

func (gen CodeGenerator) methodCallCountGetter(method *ast.Field) *ast.FuncDecl {
	return &ast.FuncDecl{
		Name: ast.NewIdent(gen.callCountMethodName(method)),
		Type: &ast.FuncType{
			Results: &ast.FieldList{List: []*ast.Field{
				&ast.Field{
					Type: ast.NewIdent("int"),
				},
			}},
		},
		Recv: gen.receiverFieldList(),
		Body: &ast.BlockStmt{List: []ast.Stmt{
			gen.callMutex(method, "RLock"),
			gen.deferMutex(method, "RUnlock"),

			&ast.ReturnStmt{
				Results: []ast.Expr{
					&ast.CallExpr{
						Fun: ast.NewIdent("len"),
						Args: []ast.Expr{
							&ast.SelectorExpr{
								X:   receiverIdent(),
								Sel: ast.NewIdent(gen.callArgsFieldName(method)),
							},
						},
					},
				},
			},
		}},
	}
}

func (gen CodeGenerator) methodCallArgsGetter(method *ast.Field) *ast.FuncDecl {
	indexIdent := ast.NewIdent("i")
	resultValues := []ast.Expr{}
	resultTypes := []*ast.Field{}

	eachMethodParam(method.Type.(*ast.FuncType), func(name string, t ast.Expr, _ int) {
		resultValues = append(resultValues, &ast.SelectorExpr{
			X: &ast.IndexExpr{
				X: &ast.SelectorExpr{
					X:   receiverIdent(),
					Sel: ast.NewIdent(gen.callArgsFieldName(method)),
				},
				Index: indexIdent,
			},
			Sel: ast.NewIdent(name),
		})

		resultTypes = append(resultTypes, &ast.Field{
			Type: storedTypeForType(t),
		})
	})

	return &ast.FuncDecl{
		Name: ast.NewIdent(gen.callArgsMethodName(method)),
		Type: &ast.FuncType{
			Params: &ast.FieldList{List: []*ast.Field{
				&ast.Field{
					Names: []*ast.Ident{indexIdent},
					Type:  ast.NewIdent("int"),
				},
			}},
			Results: &ast.FieldList{List: resultTypes},
		},
		Recv: gen.receiverFieldList(),
		Body: &ast.BlockStmt{List: []ast.Stmt{
			gen.callMutex(method, "RLock"),
			gen.deferMutex(method, "RUnlock"),
			&ast.ReturnStmt{
				Results: resultValues,
			},
		}},
	}
}

func (gen CodeGenerator) methodReturnsSetter(method *ast.Field) *ast.FuncDecl {
	methodType := method.Type.(*ast.FuncType)

	params := []*ast.Field{}
	structFields := []ast.Expr{}
	eachMethodResult(methodType, func(name string, t ast.Expr) {
		params = append(params, &ast.Field{
			Names: []*ast.Ident{ast.NewIdent(name)},
			Type:  t,
		})

		structFields = append(structFields, ast.NewIdent(name))
	})

	return &ast.FuncDecl{
		Name: ast.NewIdent(gen.returnSetterMethodName(method)),
		Type: &ast.FuncType{
			Params: &ast.FieldList{List: params},
		},
		Recv: gen.receiverFieldList(),
		Body: &ast.BlockStmt{List: []ast.Stmt{
			&ast.AssignStmt{
				Tok: token.ASSIGN,
				Lhs: []ast.Expr{
					&ast.SelectorExpr{
						X:   receiverIdent(),
						Sel: ast.NewIdent(gen.methodStubFuncName(method)),
					},
				},
				Rhs: []ast.Expr{
					&ast.BasicLit{
						Kind:  token.STRING,
						Value: "nil",
					},
				},
			},
			&ast.AssignStmt{
				Tok: token.ASSIGN,
				Lhs: []ast.Expr{
					&ast.SelectorExpr{
						X:   receiverIdent(),
						Sel: ast.NewIdent(gen.returnStructFieldName(method)),
					},
				},
				Rhs: []ast.Expr{
					&ast.CompositeLit{
						Type: returnStructTypeForMethod(methodType),
						Elts: structFields,
					},
				},
			},
		}},
	}
}

func (gen CodeGenerator) methodReturnsOnCallSetter(method *ast.Field) *ast.FuncDecl {
	methodType := method.Type.(*ast.FuncType)

	params := []*ast.Field{}
	structFields := []ast.Expr{}
	eachMethodResult(methodType, func(name string, t ast.Expr) {
		params = append(params, &ast.Field{
			Names: []*ast.Ident{ast.NewIdent(name)},
			Type:  t,
		})

		structFields = append(structFields, ast.NewIdent(name))
	})

	return &ast.FuncDecl{
		Name: ast.NewIdent(gen.returnSetterOnCallMethodName(method)),
		Type: &ast.FuncType{
			Params: &ast.FieldList{List: append([]*ast.Field{&ast.Field{
				Names: []*ast.Ident{ast.NewIdent("i")},
				Type:  ast.NewIdent("int"),
			}}, params...)},
		},
		Recv: gen.receiverFieldList(),
		Body: &ast.BlockStmt{List: []ast.Stmt{
			&ast.AssignStmt{
				Tok: token.ASSIGN,
				Lhs: []ast.Expr{
					&ast.SelectorExpr{
						X:   receiverIdent(),
						Sel: ast.NewIdent(gen.methodStubFuncName(method)),
					},
				},
				Rhs: []ast.Expr{
					&ast.BasicLit{
						Kind:  token.STRING,
						Value: "nil",
					},
				},
			},
			&ast.IfStmt{
				Cond: nilCheck(&ast.SelectorExpr{
					X:   receiverIdent(),
					Sel: ast.NewIdent(gen.returnMapFieldName(method)),
				}),
				Body: &ast.BlockStmt{List: []ast.Stmt{
					&ast.AssignStmt{
						Tok: token.ASSIGN,
						Lhs: []ast.Expr{
							&ast.SelectorExpr{
								X:   receiverIdent(),
								Sel: ast.NewIdent(gen.returnMapFieldName(method)),
							},
						},
						Rhs: []ast.Expr{
							&ast.CallExpr{
								Fun: ast.NewIdent("make"),
								Args: []ast.Expr{
									&ast.MapType{
										Key: ast.NewIdent("int"),
										Value: &ast.StructType{
											Fields: &ast.FieldList{
												List: params,
											},
										},
									},
								},
							},
						},
					},
				}},
			},
			&ast.AssignStmt{
				Tok: token.ASSIGN,
				Lhs: []ast.Expr{
					&ast.IndexExpr{
						X: &ast.SelectorExpr{
							X:   receiverIdent(),
							Sel: ast.NewIdent(gen.returnMapFieldName(method)),
						},
						Index: ast.NewIdent("i"),
					},
				},
				Rhs: []ast.Expr{
					&ast.CompositeLit{
						Type: returnStructTypeForMethod(methodType),
						Elts: structFields,
					},
				},
			},
		}},
	}
}

func (gen CodeGenerator) recordedInvocationsMethod() *ast.FuncDecl {
	funcNode := &ast.FuncDecl{
		Name: ast.NewIdent("Invocations"),
		Type: &ast.FuncType{
			Params: &ast.FieldList{},
			Results: &ast.FieldList{
				List: []*ast.Field{{
					Type: ast.NewIdent("map[string][][]interface{}"),
				}},
			},
		},
		Recv: gen.receiverFieldList(),
		Body: &ast.BlockStmt{List: []ast.Stmt{}},
	}

	statements := []ast.Stmt{
		&ast.ExprStmt{
			X: &ast.CallExpr{
				Fun: &ast.SelectorExpr{
					X: &ast.SelectorExpr{
						X:   receiverIdent(),
						Sel: ast.NewIdent("invocationsMutex"),
					},
					Sel: ast.NewIdent("RLock"),
				},
			},
		},
		&ast.DeferStmt{
			Call: &ast.CallExpr{
				Fun: &ast.SelectorExpr{
					X: &ast.SelectorExpr{
						X:   receiverIdent(),
						Sel: ast.NewIdent("invocationsMutex"),
					},
					Sel: ast.NewIdent("RUnlock"),
				},
			},
		},
	}

	returnStmt := &ast.ReturnStmt{
		Results: []ast.Expr{
			&ast.SelectorExpr{
				X:   receiverIdent(),
				Sel: ast.NewIdent("invocations"),
			},
		},
	}

	for _, m := range gen.Model.Methods {
		methodMutexFieldName := gen.mutexFieldName(m.Field)
		lockStmt := &ast.ExprStmt{
			X: &ast.CallExpr{
				Fun: &ast.SelectorExpr{
					X: &ast.SelectorExpr{
						X:   receiverIdent(),
						Sel: ast.NewIdent(methodMutexFieldName),
					},
					Sel: ast.NewIdent("RLock"),
				},
			},
		}
		unlockStmt := &ast.DeferStmt{
			Call: &ast.CallExpr{
				Fun: &ast.SelectorExpr{
					X: &ast.SelectorExpr{
						X:   receiverIdent(),
						Sel: ast.NewIdent(methodMutexFieldName),
					},
					Sel: ast.NewIdent("RUnlock"),
				},
			},
		}

		statements = append(statements, lockStmt)
		statements = append(statements, unlockStmt)
	}

	funcNode.Body.List = append(statements, returnStmt)
	return funcNode
}

func (gen CodeGenerator) recordInvocationMethod() *ast.FuncDecl {
	return &ast.FuncDecl{
		Name: ast.NewIdent("recordInvocation"),
		Type: &ast.FuncType{
			Params: &ast.FieldList{
				List: []*ast.Field{{
					Names: []*ast.Ident{ast.NewIdent("key")},
					Type:  ast.NewIdent("string"),
				},
					{
						Names: []*ast.Ident{ast.NewIdent("args")},
						Type:  ast.NewIdent("[]interface{}"),
					}},
			},
			Results: &ast.FieldList{},
		},
		Recv: gen.receiverFieldList(),
		Body: &ast.BlockStmt{List: []ast.Stmt{
			&ast.ExprStmt{
				X: &ast.CallExpr{
					Fun: &ast.SelectorExpr{
						X: &ast.SelectorExpr{
							X:   receiverIdent(),
							Sel: ast.NewIdent("invocationsMutex"),
						},
						Sel: ast.NewIdent("Lock"),
					},
				},
			},
			&ast.DeferStmt{
				Call: &ast.CallExpr{
					Fun: &ast.SelectorExpr{
						X: &ast.SelectorExpr{
							X:   receiverIdent(),
							Sel: ast.NewIdent("invocationsMutex"),
						},
						Sel: ast.NewIdent("Unlock"),
					},
				},
			},
			&ast.IfStmt{
				Cond: nilCheck(&ast.SelectorExpr{
					X:   receiverIdent(),
					Sel: ast.NewIdent("invocations"),
				}),
				Body: &ast.BlockStmt{
					List: []ast.Stmt{
						&ast.AssignStmt{
							Tok: token.ASSIGN,
							Lhs: []ast.Expr{&ast.SelectorExpr{
								X:   receiverIdent(),
								Sel: ast.NewIdent("invocations"),
							}},
							Rhs: []ast.Expr{ast.NewIdent("map[string][][]interface{}{}")},
						},
					},
				},
			},

			&ast.IfStmt{
				Cond: nilCheck(&ast.IndexExpr{
					X: &ast.SelectorExpr{
						X:   receiverIdent(),
						Sel: ast.NewIdent("invocations"),
					},
					Index: ast.NewIdent("key"),
				}),
				Body: &ast.BlockStmt{
					List: []ast.Stmt{
						&ast.AssignStmt{
							Tok: token.ASSIGN,
							Lhs: []ast.Expr{&ast.IndexExpr{
								X: &ast.SelectorExpr{
									X:   receiverIdent(),
									Sel: ast.NewIdent("invocations"),
								},
								Index: ast.NewIdent("key"),
							}},
							Rhs: []ast.Expr{ast.NewIdent("[][]interface{}{}")},
						},
					},
				},
			},

			&ast.AssignStmt{
				Tok: token.ASSIGN,
				Lhs: []ast.Expr{&ast.IndexExpr{
					X: &ast.SelectorExpr{
						X:   receiverIdent(),
						Sel: ast.NewIdent("invocations"),
					},
					Index: ast.NewIdent("key"),
				}},
				Rhs: []ast.Expr{&ast.CallExpr{
					Fun: ast.NewIdent("append"),
					Args: []ast.Expr{
						&ast.IndexExpr{
							X: &ast.SelectorExpr{
								X:   receiverIdent(),
								Sel: ast.NewIdent("invocations"),
							},
							Index: ast.NewIdent("key"),
						},
						ast.NewIdent("args"),
					},
				}},
			},
		}},
	}
}

func (gen CodeGenerator) receiverFieldList() *ast.FieldList {
	return &ast.FieldList{
		List: []*ast.Field{
			{
				Names: []*ast.Ident{receiverIdent()},
				Type:  &ast.StarExpr{X: ast.NewIdent(gen.StructName)},
			},
		},
	}
}

func (gen CodeGenerator) ensureInterfaceIsUsed() *ast.GenDecl {
	packageName := gen.packageAlias[strconv.Quote(gen.Model.ImportPath)]
	if gen.Model.RepresentedByInterface {
		return &ast.GenDecl{
			Tok: token.VAR,
			Specs: []ast.Spec{
				&ast.ValueSpec{
					Names: []*ast.Ident{ast.NewIdent("_")},
					Type: &ast.SelectorExpr{
						X:   ast.NewIdent(packageName),
						Sel: ast.NewIdent(gen.Model.Name),
					},
					Values: []ast.Expr{
						&ast.CallExpr{
							Fun:  ast.NewIdent("new"),
							Args: []ast.Expr{ast.NewIdent(gen.StructName)},
						},
					},
				},
			},
		}
	}

	return &ast.GenDecl{
		Tok: token.VAR,
		Specs: []ast.Spec{
			&ast.ValueSpec{
				Names: []*ast.Ident{ast.NewIdent("_")},
				Type: &ast.SelectorExpr{
					X:   ast.NewIdent(packageName),
					Sel: ast.NewIdent(gen.Model.Name),
				},
				Values: []ast.Expr{
					&ast.SelectorExpr{
						Sel: ast.NewIdent("Spy"),
						X: &ast.CallExpr{
							Fun:  ast.NewIdent("new"),
							Args: []ast.Expr{ast.NewIdent(gen.StructName)},
						},
					},
				},
			},
		},
	}
}

func eachMethodParam(methodType *ast.FuncType, cb func(string, ast.Expr, int)) {
	i := 0
	for _, field := range methodType.Params.List {
		if len(field.Names) == 0 {
			cb(fmt.Sprintf("arg%d", i+1), field.Type, i)
			i++
		} else {
			for _, name := range field.Names {
				cb(name.Name, field.Type, i)
				i++
			}
		}
	}
}

func eachMethodResult(methodType *ast.FuncType, cb func(string, ast.Expr)) {
	i := 1
	for _, field := range methodType.Results.List {
		if len(field.Names) == 0 {
			cb(fmt.Sprintf("result%d", i), field.Type)
			i++
		} else {
			for _ = range field.Names {
				cb(fmt.Sprintf("result%d", i), field.Type)
				i++
			}
		}
	}
}

func argsStructTypeForMethod(methodType *ast.FuncType) *ast.StructType {
	fields := []*ast.Field{}

	eachMethodParam(methodType, func(name string, t ast.Expr, _ int) {
		fields = append(fields, &ast.Field{
			Type:  storedTypeForType(t),
			Names: []*ast.Ident{ast.NewIdent(name)},
		})
	})

	return &ast.StructType{
		Fields: &ast.FieldList{List: fields},
	}
}

func returnStructTypeForMethod(methodType *ast.FuncType) *ast.StructType {
	resultFields := []*ast.Field{}
	eachMethodResult(methodType, func(name string, t ast.Expr) {
		resultFields = append(resultFields, &ast.Field{
			Type:  t,
			Names: []*ast.Ident{ast.NewIdent(name)},
		})
	})

	return &ast.StructType{
		Fields: &ast.FieldList{List: resultFields},
	}
}

func returnMapTypeForMethod(methodType *ast.FuncType) *ast.MapType {
	resultFields := []*ast.Field{}
	eachMethodResult(methodType, func(name string, t ast.Expr) {
		resultFields = append(resultFields, &ast.Field{
			Type:  t,
			Names: []*ast.Ident{ast.NewIdent(name)},
		})
	})

	return &ast.MapType{
		Key: ast.NewIdent("int"),
		Value: &ast.StructType{
			Fields: &ast.FieldList{
				List: resultFields,
			},
		},
	}
}

func storedTypeForType(t ast.Expr) ast.Expr {
	if ellipsis, ok := t.(*ast.Ellipsis); ok {
		return &ast.ArrayType{Elt: ellipsis.Elt}
	} else {
		return t
	}
}

func (gen CodeGenerator) callCountMethodName(method *ast.Field) string {
	if gen.Model.RepresentedByInterface {
		return method.Names[0].Name + "CallCount"
	} else {
		return "CallCount"
	}
}

func (gen CodeGenerator) callArgsMethodName(method *ast.Field) string {
	if gen.Model.RepresentedByInterface {
		return method.Names[0].Name + "ArgsForCall"
	} else {
		return "ArgsForCall"
	}
}

func (gen CodeGenerator) callArgsFieldName(method *ast.Field) string {
	return privatize(gen.callArgsMethodName(method))
}

func (gen CodeGenerator) mutexFieldName(method *ast.Field) string {
	if gen.Model.RepresentedByInterface {
		return privatize(method.Names[0].Name) + "Mutex"
	} else {
		return "mutex"
	}
}

func (gen CodeGenerator) methodStubFuncName(method *ast.Field) string {
	if gen.Model.RepresentedByInterface {
		return method.Names[0].Name + "Stub"
	} else {
		return "Stub"
	}
}

func (gen CodeGenerator) returnSetterMethodName(method *ast.Field) string {
	if gen.Model.RepresentedByInterface {
		return method.Names[0].Name + "Returns"
	} else {
		return "Returns"
	}
}

func (gen CodeGenerator) returnSetterOnCallMethodName(method *ast.Field) string {
	if gen.Model.RepresentedByInterface {
		return method.Names[0].Name + "ReturnsOnCall"
	} else {
		return "ReturnsOnCall"
	}
}

func (gen CodeGenerator) returnMapFieldName(method *ast.Field) string {
	return privatize(gen.returnSetterOnCallMethodName(method))
}

func (gen CodeGenerator) returnStructFieldName(method *ast.Field) string {
	return privatize(gen.returnSetterMethodName(method))
}

func receiverIdent() *ast.Ident {
	return ast.NewIdent("fake")
}

func (gen CodeGenerator) callMutex(method *ast.Field, verb string) ast.Stmt {
	return &ast.ExprStmt{
		X: &ast.CallExpr{
			Fun: &ast.SelectorExpr{
				X: &ast.SelectorExpr{
					X:   receiverIdent(),
					Sel: ast.NewIdent(gen.mutexFieldName(method)),
				},
				Sel: ast.NewIdent(verb),
			},
		},
	}
}

func (gen CodeGenerator) deferMutex(method *ast.Field, verb string) ast.Stmt {
	return &ast.DeferStmt{
		Call: &ast.CallExpr{
			Fun: &ast.SelectorExpr{
				X: &ast.SelectorExpr{
					X:   receiverIdent(),
					Sel: ast.NewIdent(gen.mutexFieldName(method)),
				},
				Sel: ast.NewIdent(verb),
			},
		},
	}
}

func publicize(input string) string {
	return strings.ToUpper(input[0:1]) + input[1:]
}

func privatize(input string) string {
	return strings.ToLower(input[0:1]) + input[1:]
}

func invertNilCheck(x ast.Expr) ast.Expr {
	return &ast.BinaryExpr{
		X:  x,
		Op: token.NEQ,
		Y: &ast.BasicLit{
			Kind:  token.STRING,
			Value: "nil",
		},
	}
}

func nilCheck(x ast.Expr) ast.Expr {
	return &ast.BinaryExpr{
		X:  x,
		Op: token.EQL,
		Y: &ast.BasicLit{
			Kind:  token.STRING,
			Value: "nil",
		},
	}
}

func quotedMethodName(method *ast.Field) *ast.Ident {
	return ast.NewIdent(`"` + method.Names[0].Name + `"`)
}

func commentLine() string {
	return "// Code generated by counterfeiter. DO NOT EDIT.\n"
}

func prettifyCode(code string) string {
	code = funcRegexp.ReplaceAllString(code, "\n\nfunc")
	code = emptyStructRegexp.ReplaceAllString(code, "struct{}")
	code = strings.Replace(code, "\n\n\n", "\n\n", -1)
	return code
}

var funcRegexp = regexp.MustCompile("\nfunc")
var emptyStructRegexp = regexp.MustCompile("struct[\\s]+{[\\s]+}")
