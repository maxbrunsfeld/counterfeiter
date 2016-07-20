package dup_packages

import "github.com/maxbrunsfeld/counterfeiter/fixtures/dup_packages/b/v1"

type DupB interface {
	B() v1.S
}
