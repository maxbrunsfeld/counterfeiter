package generator

import (
	"go/types"
	"strings"

	"golang.org/x/tools/imports"
)

// Imports indexes imports by package path and alias so that all imports have a
// unique alias, and no package is included twice.
type Imports struct {
	ByAlias   map[string]Import
	ByPkgPath map[string]Import
}

func newImports() Imports {
	return Imports{
		ByAlias:   make(map[string]Import),
		ByPkgPath: make(map[string]Import),
	}
}

// Import is a package import with the associated alias for that package.
type Import struct {
	Alias   string
	PkgPath string
}

// AddImport creates an import with the given alias and path, and adds it to
// Fake.Imports.
func (i *Imports) Add(alias string, path string) Import {
	// TODO: why is there extra whitespace on these args?
	path = imports.VendorlessPath(strings.TrimSpace(path))
	alias = strings.TrimSpace(alias)

	imp, exists := i.ByPkgPath[path]
	if exists {
		return imp
	}

	imp, exists = i.ByAlias[alias]
	if exists {
		alias = uniqueAliasForImport(alias, i.ByAlias)
	}

	result := Import{Alias: alias, PkgPath: path}
	i.ByPkgPath[path] = result
	i.ByAlias[alias] = result
	return result
}

func uniqueAliasForImport(alias string, imports map[string]Import) string {
	for i := 0; ; i++ {
		newAlias := alias + string('a'+byte(i))
		if _, exists := imports[newAlias]; !exists {
			return newAlias
		}
	}
}

// AliasForPackage returns a package alias for the package.
func (i *Imports) AliasForPackage(p *types.Package) string {
	return i.ByPkgPath[imports.VendorlessPath(p.Path())].Alias
}
