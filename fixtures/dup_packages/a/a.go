package a

import "github.com/maxbrunsfeld/counterfeiter/fixtures/dup_packages/a/v1"

type A interface {
	V1() v1.I
}
