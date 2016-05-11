package serviceadapter_test

import (
	"encoding/json"

	"github.com/pivotal-cf/on-demand-service-broker-sdk/serviceadapter"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var (
	validServiceReleasesInfo serviceadapter.ServiceReleasesInfo
)

var _ = Describe("ServiceReleasesInfo", func() {
	BeforeEach(func() {
		validServiceReleasesInfo = serviceadapter.ServiceReleasesInfo{
			DeploymentName: "service-instance-deployment",
			Releases: serviceadapter.ServiceReleases{
				{
					Name:    "release-name",
					Version: "release-version",
					Jobs:    []string{"job_one", "job_two"},
				},
			},
			Stemcell: serviceadapter.StemcellInfo{
				OS:      "BeOS",
				Version: "2",
			},
		}
	})

	Describe("(De)serialising JSON", func() {

		var expectedServiceReleasesInfo serviceadapter.ServiceReleasesInfo

		serviceReleasesInfoJSON := []byte(`{
      "deployment_name": "service-instance-deployment",
      "releases": [{
        "name": "release-name",
        "version": "release-version",
        "jobs": [
          "job_one",
          "job_two"
        ]
      }],
      "stemcell": {
        "stemcell_os": "BeOS",
        "stemcell_version": "2"
      }
    }`)

		JustBeforeEach(func() {
			expectedServiceReleasesInfo = validServiceReleasesInfo
		})

		It("deserialises a ServiceReleasesInfo object from JSON", func() {
			var serviceReleasesInfo serviceadapter.ServiceReleasesInfo
			Expect(json.Unmarshal(serviceReleasesInfoJSON, &serviceReleasesInfo)).To(Succeed())
			Expect(serviceReleasesInfo).To(Equal(validServiceReleasesInfo))
		})

		It("serialises a ServiceReleasesInfo object to JSON", func() {
			Expect(json.Marshal(expectedServiceReleasesInfo)).To(MatchJSON(serviceReleasesInfoJSON))
		})
	})

	Describe("validation", func() {
		It("returns no error when all fields non-empty", func() {
			Expect(validServiceReleasesInfo.Validate()).To(Succeed())
		})

		It("returns an error when a field is empty", func() {
			invalidServiceReleasesInfo := serviceadapter.ServiceReleasesInfo{
				DeploymentName: "service-instance-deployment",
				Stemcell: serviceadapter.StemcellInfo{
					OS:      "BeOS",
					Version: "2",
				},
			}
			Expect(invalidServiceReleasesInfo.Validate()).NotTo(Succeed())
		})
	})
})
