package the_aliased_package

//go:generate counterfeiter . InAliasedPackage
type InAliasedPackage interface {
	Stuff(int) string
}
