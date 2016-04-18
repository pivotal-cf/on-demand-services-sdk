package serviceadapter_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestServiceadapter(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Service adapter Suite")
}
