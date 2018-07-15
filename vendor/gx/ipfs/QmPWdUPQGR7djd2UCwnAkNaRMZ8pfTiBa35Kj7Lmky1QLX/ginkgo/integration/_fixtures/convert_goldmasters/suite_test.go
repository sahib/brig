package tmp_test

import (
	"testing"

	. "github.com/onsi/gomega"
	. "gx/ipfs/QmPWdUPQGR7djd2UCwnAkNaRMZ8pfTiBa35Kj7Lmky1QLX/ginkgo"
)

func TestConvertFixtures(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "ConvertFixtures Suite")
}
