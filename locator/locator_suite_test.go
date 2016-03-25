package locator_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestLocator(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Locator Suite")
}
