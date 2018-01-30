package dup_packages

import "github.com/maxbrunsfeld/counterfeiter/fixtures/dup_packages/a/v1"

//go:generate counterfeiter . DupA
type DupA interface {
	A() v1.S
}
