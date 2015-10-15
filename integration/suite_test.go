package integration_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestCounterfeiterCLIIntegration(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Counterfeiter CLI Integration Suite")
}
