package generator

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/token"
	"regexp"
	"strings"
	"unicode"

	"github.com/maxbrunsfeld/counterfeiter/model"

	"golang.org/x/tools/imports"
)

type CodeGenerator struct {
	Model       model.InterfaceToFake
	StructName  string
	PackageName string
}

func (gen CodeGenerator) GenerateFake() (string, error) {
	buf := new(bytes.Buffer)
	err := format.Node(buf, token.NewFileSet(), gen.sourceFile())
	if err != nil {
		return "", err
	}

	code, err := imports.Process("", buf.Bytes(), nil)
	return commentLine() + prettifyCode(string(code)), err
}

func (gen CodeGenerator) isExportedInterface() bool {
	return unicode.IsUpper([]rune(gen.Model.Name)[0])
}

func (gen CodeGenerator) sourceFile() ast.Node {
	declarations := []ast.Decl{
		gen.imports(),
		gen.fakeStructType(),
	}

	for _, method := range gen.Model.Methods {
		methodType := method.Type.(*ast.FuncType)

		declarations = append(
			declarations,
			gen.stubbedMethodImplementation(method),
			gen.methodCallCountGetter(method),
		)

		if methodType.Params.NumFields() > 0 {
			declarations = append(
				declarations,
				gen.methodCallArgsGetter(method),
			)
		}

		if methodType.Results.NumFields() > 0 {
			declarations = append(
				declarations,
				gen.methodReturnsSetter(method),
			)
		}
	}

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
	specs := []ast.Spec{
		&ast.ImportSpec{
			Path: &ast.BasicLit{
				Kind:  token.STRING,
				Value: `"` + gen.Model.ImportPath + `"`,
			},
		},
		&ast.ImportSpec{
			Path: &ast.BasicLit{
				Kind:  token.STRING,
				Value: `"sync"`,
			},
		},
	}

	for _, spec := range gen.Model.ImportSpecs {
		specs = append(specs, spec)
	}

	return &ast.GenDecl{
		Lparen: 1,
		Tok:    token.IMPORT,
		Specs:  specs,
	}
}

func (gen CodeGenerator) fakeStructType() ast.Decl {
	structFields := []*ast.Field{}

	for _, method := range gen.Model.Methods {
		methodType := method.Type.(*ast.FuncType)

		structFields = append(
			structFields,

			&ast.Field{
				Names: []*ast.Ident{ast.NewIdent(gen.methodStubFuncName(method))},
				Type:  method.Type,
			},

			&ast.Field{
				Type: &ast.SelectorExpr{
					X:   ast.NewIdent("sync"),
					Sel: ast.NewIdent("RWMutex"),
				},
				Names: []*ast.Ident{ast.NewIdent(gen.mutexFieldName(method))},
			},

			&ast.Field{
				Names: []*ast.Ident{ast.NewIdent(gen.callArgsFieldName(method))},
				Type: &ast.ArrayType{
					Elt: argsStructTypeForMethod(methodType),
				},
			},
		)

		if methodType.Results.NumFields() > 0 {
			structFields = append(
				structFields,
				&ast.Field{
					Names: []*ast.Ident{ast.NewIdent(gen.returnStructFieldName(method))},
					Type:  returnStructTypeForMethod(methodType),
				},
			)
		}
	}

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
				&ast.AssignStmt{
					Tok: token.DEFINE,
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

	var lastStatement ast.Stmt
	if methodType.Results.NumFields() > 0 {
		returnValues := []ast.Expr{}
		eachMethodResult(methodType, func(name string, t ast.Expr) {
			returnValues = append(returnValues, &ast.SelectorExpr{
				X: &ast.SelectorExpr{
					X:   receiverIdent(),
					Sel: ast.NewIdent(gen.returnStructFieldName(method)),
				},
				Sel: ast.NewIdent(name),
			})
		})

		lastStatement = &ast.IfStmt{
			Cond: nilCheck(stubFunc),
			Body: &ast.BlockStmt{List: []ast.Stmt{
				&ast.ReturnStmt{Results: []ast.Expr{stubFuncCall}},
			}},
			Else: &ast.BlockStmt{List: []ast.Stmt{
				&ast.ReturnStmt{Results: returnValues},
			}},
		}
	} else {
		lastStatement = &ast.IfStmt{
			Cond: nilCheck(stubFunc),
			Body: &ast.BlockStmt{List: []ast.Stmt{
				&ast.ExprStmt{X: stubFuncCall},
			}},
		}
	}

	bodyStatements = append(bodyStatements,
		gen.callMutex(method, "Lock"),

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

		gen.callMutex(method, "Unlock"),

		lastStatement,
	)

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
	if gen.Model.RepresentedByInterface {
		return &ast.GenDecl{
			Tok: token.VAR,
			Specs: []ast.Spec{
				&ast.ValueSpec{
					Names: []*ast.Ident{ast.NewIdent("_")},
					Type: &ast.SelectorExpr{
						X:   ast.NewIdent(gen.Model.PackageName),
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
	} else {
		return &ast.GenDecl{
			Tok: token.VAR,
			Specs: []ast.Spec{
				&ast.ValueSpec{
					Names: []*ast.Ident{ast.NewIdent("_")},
					Type: &ast.SelectorExpr{
						X:   ast.NewIdent(gen.Model.PackageName),
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
	for i, field := range methodType.Results.List {
		cb(fmt.Sprintf("result%d", i+1), field.Type)
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

func nilCheck(x ast.Expr) ast.Expr {
	return &ast.BinaryExpr{
		X:  x,
		Op: token.NEQ,
		Y: &ast.BasicLit{
			Kind:  token.STRING,
			Value: "nil",
		},
	}
}

func commentLine() string {
	return "// This file was generated by counterfeiter\n"
}

func prettifyCode(code string) string {
	code = funcRegexp.ReplaceAllString(code, "\n\nfunc")
	code = emptyStructRegexp.ReplaceAllString(code, "struct{}")
	code = strings.Replace(code, "\n\n\n", "\n\n", -1)
	return code
}

var funcRegexp = regexp.MustCompile("\nfunc")
var emptyStructRegexp = regexp.MustCompile("struct[\\s]+{[\\s]+}")
