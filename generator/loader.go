package generator

import (
	"fmt"
	"go/types"
	"log"
	"reflect"

	"golang.org/x/tools/go/packages"
)

func (f *Fake) loadPackages(packagePath string) error {
	p, err := packages.Load(&packages.Config{
		Mode: packages.LoadSyntax,
	}, packagePath)
	if err != nil {
		return err
	}
	f.Packages = p
	return nil
}

func (f *Fake) findPackageWithTarget() error {
	var target *types.TypeName
	var pkg *packages.Package
	for i := range f.Packages {
		if f.Packages[i].Types == nil || f.Packages[i].Types.Scope() == nil {
			continue
		}
		pkg = f.Packages[i]

		raw := pkg.Types.Scope().Lookup(f.TargetName)
		if raw != nil {
			if typeName, ok := raw.(*types.TypeName); ok {
				target = typeName
				break
			}
		}
	}
	if pkg == nil || target == nil {
		return fmt.Errorf("cannot find package with interface %s", f.TargetName)
	}
	f.Target = target
	f.Package = pkg
	f.TargetName = target.Name()
	f.TargetPackage = unvendor(pkg.PkgPath)
	f.TargetAlias = pkg.Name
	f.AddImport(f.TargetAlias, f.TargetPackage)
	if f.IsInterface() {
		log.Printf("Found interface with name: [%s]\n", f.TargetName)
	}
	if f.IsFunction() {
		log.Printf("Found function with name: [%s]\n", f.TargetName)
	}
	return nil
}

// addImportsFor inspects the given type and adds imports to the fake if importable
// types are found.
func (f *Fake) addImportsFor(typ types.Type) {
	if typ == nil {
		return
	}

	switch t := typ.(type) {
	case *types.Basic:
		return
	case *types.Pointer:
		f.addImportsFor(t.Elem())
	case *types.Map:
		f.addImportsFor(t.Key())
		f.addImportsFor(t.Elem())
	case *types.Chan:
		f.addImportsFor(t.Elem())
	case *types.Named:
		if t.Obj() != nil && t.Obj().Pkg() != nil {
			f.AddImport(t.Obj().Pkg().Name(), t.Obj().Pkg().Path())
		}
	case *types.Slice:
		f.addImportsFor(t.Elem())
	case *types.Array:
		f.addImportsFor(t.Elem())
	case *types.Interface:
		return
	default:
		log.Printf("!!! WARNING: Missing case for type %s\n", reflect.TypeOf(typ).String())
	}
}

func typeFor(typ types.Type, importsMap map[string]Import) string {
	if typ == nil {
		return ""
	}
	return types.TypeString(typ, func(p *types.Package) string {
		imp, ok := importsMap[unvendor(p.Path())]
		if ok {
			return imp.Alias
		}
		return ""
	})
}
