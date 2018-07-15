package passing_before_suite_test

import (
	. "github.com/onsi/gomega"
	. "gx/ipfs/QmPWdUPQGR7djd2UCwnAkNaRMZ8pfTiBa35Kj7Lmky1QLX/ginkgo"
)

var _ = Describe("PassingSuiteSetup", func() {
	It("should pass", func() {
		Ω(a).Should(Equal("ran before suite"))
		Ω(b).Should(BeEmpty())
	})

	It("should pass", func() {
		Ω(a).Should(Equal("ran before suite"))
		Ω(b).Should(BeEmpty())
	})

	It("should pass", func() {
		Ω(a).Should(Equal("ran before suite"))
		Ω(b).Should(BeEmpty())
	})

	It("should pass", func() {
		Ω(a).Should(Equal("ran before suite"))
		Ω(b).Should(BeEmpty())
	})
})
