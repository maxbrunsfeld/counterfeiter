// +build go1.11

package integration_test

import (
	"testing"

	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestIntegration(t *testing.T) {
	suite := spec.New("integration", spec.Report(report.Terminal{}))
	suite("round trip as module", testRoundTripAsModule)
	suite("round trip with gopath", testRoundTrip)
	suite.Run(t)
}

func testRoundTripAsModule(t *testing.T, when spec.G, it spec.S) {
	runTests(false, t, when, it)
}

func testRoundTrip(t *testing.T, when spec.G, it spec.S) {
	runTests(true, t, when, it)
}
