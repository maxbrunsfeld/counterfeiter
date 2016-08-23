package generator

import (
	"bytes"
	"go/ast"
	"go/format"
	"go/token"

	"os"
	"strings"

	"path/filepath"

	"github.com/maxbrunsfeld/counterfeiter/model"
	"golang.org/x/tools/imports"
)

type InterfaceGenerator struct {
	Model                  *model.PackageToInterfacify
	Package                string
	DestinationInterface   string
	DestinationPackageName string
}

func (ig InterfaceGenerator) GenerateInterface() (string, error) {

	buf := new(bytes.Buffer)
	err := format.Node(buf, token.NewFileSet(), ig.outputAST())
	if err != nil {
		return "", err
	}
	code, err := imports.Process("", buf.Bytes(), nil)
	if err != nil {
		panic(err)
		return "", err
	}

	return commentLine() + "// with command: counterfeiter " + strings.Join(os.Args[1:], " ") + "\n" + prettifyCode(string(code)), nil
}

func (ig InterfaceGenerator) outputAST() *ast.File {

	declarations := []ast.Decl{}

	declarations = append(declarations, ig.interfaceDecl())

	return &ast.File{
		Name:  &ast.Ident{Name: ig.DestinationPackageName},
		Decls: declarations,
	}
}

func (ig InterfaceGenerator) interfaceDecl() *ast.GenDecl {
	fakeFilePath := filepath.Join(strings.ToLower(ig.Model.Name)+"_fake", "fake_"+strings.ToLower(ig.Model.Name)+".go")
	return &ast.GenDecl{
		Tok: token.TYPE,
		Doc: &ast.CommentGroup{[]*ast.Comment{{Text: "//go:generate counterfeiter -o " + fakeFilePath + " . " + ig.DestinationInterface}}},
		Specs: []ast.Spec{
			&ast.TypeSpec{
				Name: &ast.Ident{Name: ig.DestinationInterface},
				Type: &ast.InterfaceType{
					Methods: ig.methods(),
				},
			},
		},
	}
}

func (ig InterfaceGenerator) methods() *ast.FieldList {
	fieldList := []*ast.Field{}

	for _, funcDecl := range ig.Model.Funcs {
		field := &ast.Field{
			Names: []*ast.Ident{funcDecl.Name},
			Type:  funcDecl.Type,
		}
		fieldList = append(fieldList, field)
	}

	return &ast.FieldList{List: fieldList}
}
