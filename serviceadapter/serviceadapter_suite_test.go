package serviceadapter_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"

	"testing"
)

var adapterBin string

var _ = BeforeSuite(func() {
	var err error
	adapterBin, err = gexec.Build("github.com/pivotal-cf/on-demand-service-broker-sdk/serviceadapter/testharness")
	Expect(err).NotTo(HaveOccurred())
})

var _ = AfterSuite(func() {
	gexec.CleanupBuildArtifacts()
})

func TestServiceadapter(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Service adapter Suite")
}
