package serviceadapter_test

import (
	"encoding/json"

	"github.com/pivotal-cf/on-demand-service-broker-sdk/serviceadapter"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var (
	validServiceReleasesInfo serviceadapter.ServiceDeployment
)

var _ = Describe("ServiceDeployment", func() {
	BeforeEach(func() {
		validServiceReleasesInfo = serviceadapter.ServiceDeployment{
			DeploymentName: "service-instance-deployment",
			Releases: serviceadapter.ServiceReleases{
				{
					Name:    "release-name",
					Version: "release-version",
					Jobs:    []string{"job_one", "job_two"},
				},
			},
			Stemcell: serviceadapter.Stemcell{
				OS:      "BeOS",
				Version: "2",
			},
		}
	})

	Describe("(De)serialising JSON", func() {

		var expectedServiceReleasesInfo serviceadapter.ServiceDeployment

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
			var serviceReleasesInfo serviceadapter.ServiceDeployment
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
			invalidServiceReleasesInfo := serviceadapter.ServiceDeployment{
				DeploymentName: "service-instance-deployment",
				Stemcell: serviceadapter.Stemcell{
					OS:      "BeOS",
					Version: "2",
				},
			}
			Expect(invalidServiceReleasesInfo.Validate()).NotTo(Succeed())
		})
	})
})
