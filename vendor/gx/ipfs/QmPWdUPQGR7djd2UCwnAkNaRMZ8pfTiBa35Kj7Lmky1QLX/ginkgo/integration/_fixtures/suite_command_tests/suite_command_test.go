package suite_command_test

import (
	. "github.com/onsi/gomega"
	. "gx/ipfs/QmPWdUPQGR7djd2UCwnAkNaRMZ8pfTiBa35Kj7Lmky1QLX/ginkgo"
)

var _ = Describe("Testing suite command", func() {
	It("it should succeed", func() {
		Ω(true).Should(Equal(true))
	})

	PIt("a failing test", func() {
		It("should fail", func() {
			Ω(true).Should(Equal(false))
		})
	})
})
