package fixtures

import "net/http"

type EmbedsInterfaces interface {
	http.Handler
	InterfaceToEmbed

	DoThings()
}

type InterfaceToEmbed interface {
	EmbeddedMethod() string
}
