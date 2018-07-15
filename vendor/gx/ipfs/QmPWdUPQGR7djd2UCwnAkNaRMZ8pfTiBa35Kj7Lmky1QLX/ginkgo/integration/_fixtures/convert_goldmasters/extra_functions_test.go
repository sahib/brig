package tmp

import (
	. "gx/ipfs/QmPWdUPQGR7djd2UCwnAkNaRMZ8pfTiBa35Kj7Lmky1QLX/ginkgo"
)

var _ = Describe("Testing with Ginkgo", func() {
	It("something less important", func() {

		strp := "hello!"
		somethingImportant(GinkgoT(), &strp)
	})
})

func somethingImportant(t GinkgoTInterface, message *string) {
	t.Log("Something important happened in a test: " + *message)
}
