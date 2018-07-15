package failing_before_suite_test

import (
	. "gx/ipfs/QmPWdUPQGR7djd2UCwnAkNaRMZ8pfTiBa35Kj7Lmky1QLX/ginkgo"
)

var _ = Describe("FailingBeforeSuite", func() {
	It("should never run", func() {
		println("NEVER SEE THIS")
	})

	It("should never run", func() {
		println("NEVER SEE THIS")
	})
})
