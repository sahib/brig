package does_not_compile_test

import (
	. "github.com/onsi/gomega"
	. "gx/ipfs/QmPWdUPQGR7djd2UCwnAkNaRMZ8pfTiBa35Kj7Lmky1QLX/ginkgo"

	"testing"
)

func TestDoes_not_compile(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Does_not_compile Suite")
}
