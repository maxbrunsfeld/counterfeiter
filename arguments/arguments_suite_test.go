package arguments_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestArguments(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Arguments Suite")
}
