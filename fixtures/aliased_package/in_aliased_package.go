package the_aliased_package // import "github.com/maxbrunsfeld/counterfeiter/v6/fixtures/aliased_package"

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . InAliasedPackage
type InAliasedPackage interface {
	Stuff(int) string
}
