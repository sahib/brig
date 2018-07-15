package more_ginkgo_tests_test

import (
	. "github.com/onsi/gomega"
	. "gx/ipfs/QmPWdUPQGR7djd2UCwnAkNaRMZ8pfTiBa35Kj7Lmky1QLX/ginkgo"
	. "gx/ipfs/QmPWdUPQGR7djd2UCwnAkNaRMZ8pfTiBa35Kj7Lmky1QLX/ginkgo/integration/_fixtures/more_ginkgo_tests"
)

var _ = Describe("MoreGinkgoTests", func() {
	It("should pass", func() {
		Ω(AlwaysTrue()).Should(BeTrue())
	})

	It("should always pass", func() {
		Ω(AlwaysTrue()).Should(BeTrue())
	})
})
