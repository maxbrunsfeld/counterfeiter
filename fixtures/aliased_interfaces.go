package fixtures

import alias "github.com/maxbrunsfeld/counterfeiter/v6/fixtures/another_package"

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . AliasedInterface
//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate
//counterfeiter:generate . AliasedInterface

// AliasedInterface is an interface that embeds an interface in an aliased package.
type AliasedInterface interface {
	alias.AnotherInterface
}
