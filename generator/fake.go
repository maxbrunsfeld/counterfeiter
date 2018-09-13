package generator

import (
	"bytes"
	"errors"
	"go/types"
	"html/template"
	"log"
	"unicode"
	"unicode/utf8"

	"golang.org/x/tools/go/packages"
	"golang.org/x/tools/imports"
)

// Fake is used to generate a Fake implementation of an interface.
type Fake struct {
	Packages           []*packages.Package
	Package            *packages.Package
	Target             *types.TypeName
	DestinationPackage string
	Name               string
	TargetAlias        string
	TargetName         string
	TargetPackage      string
	Imports            []Import
	Methods            []Method
	Method             Method
}

// Method is a method of the interface.
type Method struct {
	FakeName string
	Name     string
	Params   Params
	Args     string
	Returns  Returns
	Rets     string
}

// NewFake returns a Fake that loads the package and finds the interface or the
// function.
func NewFake(interfaceName string, packagePath string, fakeName string, destinationPackage string) (*Fake, error) {
	f := &Fake{
		TargetName:         interfaceName,
		TargetPackage:      packagePath,
		Name:               fakeName,
		DestinationPackage: destinationPackage,
		Imports: []Import{
			Import{
				Alias: "sync",
				Path:  "sync",
			},
		},
	}

	err := f.loadPackages(packagePath)
	if err != nil {
		return nil, err
	}

	err = f.findPackageWithTarget()
	if err != nil {
		return nil, err
	}

	if f.IsInterface() {
		f.loadMethodsForInterface()
	}
	if f.IsFunction() {
		err := f.loadMethodForFunction()
		if err != nil {
			return nil, err
		}
	}
	return f, nil
}

// IsInterface indicates whether the fake is for an interface.
func (f *Fake) IsInterface() bool {
	if f.Target == nil {
		return false
	}
	return types.IsInterface(f.Target.Type())
}

// IsFunction indicates whether the fake is for a function..
func (f *Fake) IsFunction() bool {
	if f.Target == nil {
		return false
	}
	return !f.IsInterface()
}

func unexport(s string) string {
	if s == "" {
		return ""
	}
	r, n := utf8.DecodeRuneInString(s)
	return string(unicode.ToLower(r)) + s[n:]
}

// Generate uses the Fake to generate an implementation, optionally running
// goimports on the output.
func (f *Fake) Generate(runImports bool) ([]byte, error) {
	var tmpl *template.Template
	if f.IsInterface() {
		log.Printf("Writing fake %s for interface %s to package %s\n", f.Name, f.TargetName, f.DestinationPackage)
		tmpl = template.Must(template.New("fake").Funcs(interfaceFuncs).Parse(interfaceTemplate))
	}
	if f.IsFunction() {
		log.Printf("Writing fake %s for function %s to package %s\n", f.Name, f.TargetName, f.DestinationPackage)
		tmpl = template.Must(template.New("fake").Funcs(functionFuncs).Parse(functionTemplate))
	}
	if tmpl == nil {
		return nil, errors.New("counterfeiter can only generate fakes for interfaces or specific functions")
	}

	b := &bytes.Buffer{}
	tmpl.Execute(b, f)
	if runImports {
		return imports.Process("counterfeiter_temp_process_file", b.Bytes(), nil)
	}
	return b.Bytes(), nil
}
