package nested_test

import (
	"testing"

	. "github.com/onsi/gomega"
	. "gx/ipfs/QmPWdUPQGR7djd2UCwnAkNaRMZ8pfTiBa35Kj7Lmky1QLX/ginkgo"
)

func TestNested(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Nested Suite")
}
