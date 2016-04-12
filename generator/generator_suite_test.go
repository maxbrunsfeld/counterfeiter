package generator_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestGenerator(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Generator Suite")
}
