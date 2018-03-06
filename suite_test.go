package main_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestCounterfeiter(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Counterfeiter Suite")
}
