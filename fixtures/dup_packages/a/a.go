package a

import "github.com/maxbrunsfeld/counterfeiter/fixtures/dup_packages/a/v1"

//go:generate counterfeiter . A
type A interface {
	V1() v1.I
}
