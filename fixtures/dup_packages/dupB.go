package dup_packages

import "github.com/maxbrunsfeld/counterfeiter/fixtures/dup_packages/b/v1"

//go:generate counterfeiter . DupB
type DupB interface {
	B() v1.S
}
