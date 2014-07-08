package fixtures

import "io"

type EmbedsInterfaces interface {
	io.Writer
	InterfaceToEmbed

	DoThings()
}

type InterfaceToEmbed interface {
	EmbeddedMethod() string
}
