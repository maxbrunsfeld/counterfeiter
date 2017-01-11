package test

import (
	ioAlias "io"
)

type Potato interface {
	Tomato(ioAlias.Reader)
}
