package test

import (
	ioAlias "io"
)

//go:generate counterfeiter . Potato
type Potato interface {
	Tomato(ioAlias.Reader)
}
