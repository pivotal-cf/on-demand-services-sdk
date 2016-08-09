package serviceadapter_test

import (
	"encoding/json"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/on-demand-service-broker-sdk/serviceadapter"
)

var _ = Describe("ServiceRelease", func() {
	Describe("(De)serialising from JSON", func() {
		serviceReleaseJson := []byte(`{
	    "name": "kafka",
	    "version": "dev.1",
	    "jobs": ["kafka_node", "zookeeper", "whatever you need"]
	  }`)

		expectedServiceRelease := serviceadapter.ServiceRelease{
			Name:    "kafka",
			Version: "dev.1",
			Jobs:    []string{"kafka_node", "zookeeper", "whatever you need"},
		}

		It("deserialises JSON into a ServiceRelease object", func() {
			var serviceRelease serviceadapter.ServiceRelease
			Expect(json.Unmarshal(serviceReleaseJson, &serviceRelease)).To(Succeed())
			Expect(serviceRelease).To(Equal(expectedServiceRelease))
		})

		It("serialises a ServiceRelease object to JSON", func() {
			Expect(toJson(expectedServiceRelease)).To(MatchJSON(serviceReleaseJson))
		})
	})

	Describe("Validating", func() {
		It("returns no error when there is at least one valid release", func() {
			serviceReleases := serviceadapter.ServiceReleases{
				{Name: "foo", Version: "bar", Jobs: []string{"baz"}},
			}
			Expect(serviceReleases.Validate()).To(Succeed())
		})

		It("returns an error if there are no service releases", func() {
			serviceReleases := serviceadapter.ServiceReleases{}
			Expect(serviceReleases.Validate()).NotTo(Succeed())
		})

		It("returns an error if a release is missing a field", func() {
			serviceReleases := serviceadapter.ServiceReleases{
				{Name: "foo", Version: "bar", Jobs: []string{"baz"}},
				{Name: "qux", Jobs: []string{"quux"}},
			}
			Expect(serviceReleases.Validate()).NotTo(Succeed())
		})

		It("returns an error if a release provides no jobs", func() {
			serviceReleases := serviceadapter.ServiceReleases{
				{Name: "foo", Version: "bar", Jobs: []string{"baz"}},
				{Name: "qux", Version: "quux", Jobs: []string{}},
			}
			Expect(serviceReleases.Validate()).NotTo(Succeed())
		})
	})
})
