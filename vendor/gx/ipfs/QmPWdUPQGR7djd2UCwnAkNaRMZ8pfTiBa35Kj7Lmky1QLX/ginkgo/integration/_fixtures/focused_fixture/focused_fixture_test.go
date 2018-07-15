package focused_fixture_test

import (
	. "gx/ipfs/QmPWdUPQGR7djd2UCwnAkNaRMZ8pfTiBa35Kj7Lmky1QLX/ginkgo"
	. "gx/ipfs/QmPWdUPQGR7djd2UCwnAkNaRMZ8pfTiBa35Kj7Lmky1QLX/ginkgo/extensions/table"
)

var _ = Describe("FocusedFixture", func() {
	FDescribe("focused", func() {
		It("focused", func() {

		})
	})

	FContext("focused", func() {
		It("focused", func() {

		})
	})

	FIt("focused", func() {

	})

	FMeasure("focused", func(b Benchmarker) {

	}, 2)

	FDescribeTable("focused",
		func() {},
		Entry("focused"),
	)

	DescribeTable("focused",
		func() {},
		FEntry("focused"),
	)

	Describe("not focused", func() {
		It("not focused", func() {

		})
	})

	Context("not focused", func() {
		It("not focused", func() {

		})
	})

	It("not focused", func() {

	})

	Measure("not focused", func(b Benchmarker) {

	}, 2)

	DescribeTable("not focused",
		func() {},
		Entry("not focused"),
	)
})
