// Copyright (C) 2016-Present Pivotal Software, Inc. All rights reserved.

// This program and the accompanying materials are made available under
// the terms of the under the Apache License, Version 2.0 (the "License”);
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at

// http://www.apache.org/licenses/LICENSE-2.0

// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package serviceadapter_test

import (
	"encoding/json"

	"github.com/pivotal-cf/on-demand-services-sdk/serviceadapter"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var (
	validDeployment serviceadapter.ServiceDeployment
)

var _ = Describe("ServiceDeployment", func() {
	BeforeEach(func() {
		validDeployment = serviceadapter.ServiceDeployment{
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

		var expectedServiceDeployment serviceadapter.ServiceDeployment

		serviceDeploymentJSON := []byte(`{
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
			expectedServiceDeployment = validDeployment
		})

		It("deserialises a ServiceDeployment object from JSON", func() {
			var serviceDeployment serviceadapter.ServiceDeployment
			Expect(json.Unmarshal(serviceDeploymentJSON, &serviceDeployment)).To(Succeed())
			Expect(serviceDeployment).To(Equal(validDeployment))
		})

		It("serialises a ServiceDeployment object to JSON", func() {
			Expect(toJson(expectedServiceDeployment)).To(MatchJSON(serviceDeploymentJSON))
		})
	})

	Describe("validation", func() {
		It("returns no error when all fields non-empty", func() {
			Expect(validDeployment.Validate()).To(Succeed())
		})

		It("returns an error when a field is empty", func() {
			invalidServiceDeployment := serviceadapter.ServiceDeployment{
				DeploymentName: "service-instance-deployment",
				Stemcell: serviceadapter.Stemcell{
					OS:      "BeOS",
					Version: "2",
				},
			}
			Expect(invalidServiceDeployment.Validate()).NotTo(Succeed())
		})
	})
})
