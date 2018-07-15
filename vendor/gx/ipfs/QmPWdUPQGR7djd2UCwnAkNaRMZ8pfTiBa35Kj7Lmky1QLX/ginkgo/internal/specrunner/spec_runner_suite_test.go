package specrunner_test

import (
	. "github.com/onsi/gomega"
	. "gx/ipfs/QmPWdUPQGR7djd2UCwnAkNaRMZ8pfTiBa35Kj7Lmky1QLX/ginkgo"

	"testing"
)

func TestSpecRunner(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Spec Runner Suite")
}
