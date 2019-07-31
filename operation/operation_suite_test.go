package operation_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestOperation(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Operation Suite")
}
