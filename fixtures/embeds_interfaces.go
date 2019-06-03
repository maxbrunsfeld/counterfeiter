package fixtures

import (
	"net/http"

	"github.com/maxbrunsfeld/counterfeiter/v6/fixtures/another_package"
)

//counterfeiter:generate . EmbedsInterfaces
type EmbedsInterfaces interface {
	http.Handler
	another_package.AnotherInterface
	InterfaceToEmbed

	DoThings()
}

type InterfaceToEmbed interface {
	EmbeddedMethod() string
}
