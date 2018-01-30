package dup_packages

//go:generate counterfeiter . DupAB
type DupAB interface {
	DupA
	DupB
}
