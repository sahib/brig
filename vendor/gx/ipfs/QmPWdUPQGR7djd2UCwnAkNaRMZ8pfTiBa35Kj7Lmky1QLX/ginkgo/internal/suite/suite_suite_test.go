package suite_test

import (
	. "github.com/onsi/gomega"
	. "gx/ipfs/QmPWdUPQGR7djd2UCwnAkNaRMZ8pfTiBa35Kj7Lmky1QLX/ginkgo"

	"testing"
)

func Test(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Suite")
}

var numBeforeSuiteRuns = 0
var numAfterSuiteRuns = 0

var _ = BeforeSuite(func() {
	numBeforeSuiteRuns++
})

var _ = AfterSuite(func() {
	numAfterSuiteRuns++
	Ω(numBeforeSuiteRuns).Should(Equal(1))
	Ω(numAfterSuiteRuns).Should(Equal(1))
})

//Fakes
type fakeTestingT struct {
	didFail bool
}

func (fakeT *fakeTestingT) Fail() {
	fakeT.didFail = true
}
