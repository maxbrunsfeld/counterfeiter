package test

import (
	git "github.com/libgit2/git2go"
)

type Potato interface {
	Tomato(git.Oid)
}
