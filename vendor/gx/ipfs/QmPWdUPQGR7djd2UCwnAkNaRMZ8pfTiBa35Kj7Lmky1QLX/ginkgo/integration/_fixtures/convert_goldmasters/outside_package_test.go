package tmp_test

import (
	. "gx/ipfs/QmPWdUPQGR7djd2UCwnAkNaRMZ8pfTiBa35Kj7Lmky1QLX/ginkgo"
)

var _ = Describe("Testing with Ginkgo", func() {
	It("something important", func() {

		whatever := &UselessStruct{}
		if whatever.ImportantField != "SECRET_PASSWORD" {
			GinkgoT().Fail()
		}
	})
})

type UselessStruct struct {
	ImportantField string
}
