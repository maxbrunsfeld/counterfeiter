package integration_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestCounterfeiterCLI(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Counterfeiter CLI Suite")
}
