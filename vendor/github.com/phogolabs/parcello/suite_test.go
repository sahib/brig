package parcello_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestEmbedo(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Parcello Suite")
}
