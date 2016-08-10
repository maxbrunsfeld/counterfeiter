package generator

import (
	"bytes"
	"go/ast"
	"go/format"
	"go/token"
	"strconv"
	"strings"

	"github.com/maxbrunsfeld/counterfeiter/astutil"
	"github.com/maxbrunsfeld/counterfeiter/model"

	"os"

	"golang.org/x/tools/imports"
)

type ShimGenerator struct {
	Model         model.InterfaceToFake
	StructName    string
	PackageName   string
	SourcePackage string

	packageAlias map[string]string
}

func (gen ShimGenerator) GenerateReal() (string, error) {
	buf := new(bytes.Buffer)
	err := format.Node(buf, token.NewFileSet(), gen.buildASTForReal())
	if err != nil {
		panic(err)
		return "", err
	}

	code, err := imports.Process("", buf.Bytes(), nil)
	return commentLine() + "// with command: counterfeiter " + strings.Join(os.Args[1:], " ") + "\n" + prettifyCode(string(code)), err
}

func (gen ShimGenerator) isExportedInterface() bool {
	return ast.IsExported(gen.Model.Name)
}

func (gen ShimGenerator) buildASTForReal() ast.Node {
	gen.packageAlias = map[string]string{}

	declarations := []ast.Decl{}
	declarations = append(declarations, gen.imports())
	gen.fixup()

	declarations = append(declarations, gen.realStructDeclaration())

	for _, m := range gen.Model.Methods {
		declarations = append(
			declarations,
			gen.shimMethodImplementation(m.Field),
		)
	}

	return &ast.File{
		Name:  &ast.Ident{Name: gen.PackageName},
		Decls: declarations,
	}
}

func (gen ShimGenerator) imports() ast.Decl {
	specs := []ast.Spec{}
	allImports := map[string]bool{}
	dotImports := map[string]bool{}

	modelImportName := strconv.Quote(gen.Model.ImportPath)
	allImports[modelImportName] = true

	syncImportName := strconv.Quote("sync")
	allImports[syncImportName] = true
	gen.packageAlias[syncImportName] = "sync"

	for _, m := range gen.Model.Methods {
		for alias, importSpec := range m.Imports {
			if alias == "." {
				dotImports[importSpec.Name.Name] = true
				gen.packageAlias[importSpec.Path.Value] = "."
			}

			allImports[importSpec.Path.Value] = true
		}
	}

	aliases := map[string]bool{}
	aliases[gen.Model.PackageName] = true
	gen.packageAlias[modelImportName] = gen.Model.PackageName
	for importName := range allImports {
		if _, found := gen.packageAlias[importName]; found {
			continue
		}

		alias := gen.generateAlias(importName, aliases)
		if alias == "" {
			panic("could not generate an alias for " + importName)
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

func (gen ShimGenerator) generateAlias(importName string, aliases map[string]bool) string {
	unqoted, err := strconv.Unquote(importName)
	if err != nil {
		panic("cannot generate alias for " + importName)
	}
	paths := strings.Split(strings.Replace(unqoted, ".", "_", -1), "/")
	alias := ""
	for i := len(paths) - 1; i >= 0; i-- {
		alias = alias + paths[i]
		if aliases[alias] == false {
			return alias
		}
	}

	return ""
}

func (gen ShimGenerator) fixup() {
	for _, m := range gen.Model.Methods {
		typ := m.Field.Type.(*ast.FuncType)
		astutil.InjectAlias(typ, m.Imports, gen.packageAlias)
	}
}

func (gen ShimGenerator) realStructDeclaration() ast.Decl {
	structFields := []*ast.Field{}

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

func (gen ShimGenerator) shimMethodImplementation(method *ast.Field) *ast.FuncDecl {
	methodType := method.Type.(*ast.FuncType)

	realFunc := &ast.SelectorExpr{
		X:   ast.NewIdent(gen.SourcePackage),
		Sel: ast.NewIdent(method.Names[0].Name),
	}

	paramValuesToPassToStub := []ast.Expr{}
	paramFields := []*ast.Field{}
	var ellipsisPos token.Pos

	eachMethodParam(methodType, func(name string, t ast.Expr, i int) {
		paramFields = append(paramFields, &ast.Field{
			Names: []*ast.Ident{ast.NewIdent(name)},
			Type:  t,
		})

		if _, ok := t.(*ast.Ellipsis); ok {
			ellipsisPos = token.Pos(i + 1)
		}

		paramValuesToPassToStub = append(paramValuesToPassToStub, ast.NewIdent(name))
	})

	stubFuncCall := &ast.CallExpr{
		Fun:      realFunc,
		Args:     paramValuesToPassToStub,
		Ellipsis: ellipsisPos,
	}

	var lastStatement ast.Stmt
	if methodType.Results.NumFields() > 0 {
		lastStatement = &ast.ReturnStmt{Results: []ast.Expr{stubFuncCall}}
	} else {
		lastStatement = &ast.ExprStmt{X: stubFuncCall}
	}

	return &ast.FuncDecl{
		Name: method.Names[0],
		Type: &ast.FuncType{
			Params:  &ast.FieldList{List: paramFields},
			Results: methodType.Results,
		},
		Recv: gen.receiverFieldList(),
		Body: &ast.BlockStmt{List: []ast.Stmt{lastStatement}},
	}
}

func (gen ShimGenerator) receiverFieldList() *ast.FieldList {
	return &ast.FieldList{
		List: []*ast.Field{
			{
				Names: []*ast.Ident{ast.NewIdent("sh")},
				Type:  &ast.StarExpr{X: ast.NewIdent(gen.StructName)},
			},
		},
	}
}
