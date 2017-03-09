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

	yaml "gopkg.in/yaml.v2"

	"github.com/pivotal-cf/on-demand-services-sdk/serviceadapter"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Domain", func() {
	booleanPointer := func(b bool) *bool {
		return &b
	}

	Describe("RequestParameters", func() {
		Context("when arbitraryParams are present", func() {
			It("can extract arbitraryParams", func() {
				params := serviceadapter.RequestParameters{"parameters": map[string]interface{}{"foo": "bar"}}
				Expect(params.ArbitraryParams()).To(Equal(map[string]interface{}{"foo": "bar"}))
			})
		})

		Context("when arbitraryParams are absent", func() {
			It("arbitrary params are empty", func() {
				params := serviceadapter.RequestParameters{"plan_id": "baz"}
				Expect(params.ArbitraryParams()).To(Equal(map[string]interface{}{}))
			})
		})
	})

	Describe("DashboardUrl", func() {
		It("serializes dashboard_url", func() {
			dashboardUrl := serviceadapter.DashboardUrl{DashboardUrl: "https://someurl.com"}
			Expect(toJson(dashboardUrl)).To(MatchJSON(`{ "dashboard_url": "https://someurl.com"}`))
		})
	})

	Describe("InstanceGroup", func() {
		Describe("YAML unmarshalling", func() {
			var (
				yamlStr []byte
				actual  serviceadapter.InstanceGroup
			)

			BeforeEach(func() {
				yamlStr = []byte("---\nname: 'Foo'\nvm_extensions: [~, null, '', foo, bar, '']")
			})

			It("unmarshals correctly", func() {
				expected := serviceadapter.InstanceGroup{
					Name:         "Foo",
					VMExtensions: []string{"foo", "bar"},
				}

				err := yaml.Unmarshal(yamlStr, &actual)

				Expect(err).NotTo(HaveOccurred())
				Expect(actual).To(Equal(expected))
			})

			Context("when vm_extensions contains empty strings", func() {
				BeforeEach(func() {
					yamlStr = []byte("---\nvm_extensions: [~, null, '', foo, bar, '']")
				})

				It("strips out the empty strings", func() {
					expected := serviceadapter.InstanceGroup{
						VMExtensions: []string{"foo", "bar"},
					}

					err := yaml.Unmarshal(yamlStr, &actual)

					Expect(err).NotTo(HaveOccurred())
					Expect(actual.VMExtensions).To(Equal(expected.VMExtensions))
				})
			})

			Context("when vm_extensions is set to null", func() {
				BeforeEach(func() {
					yamlStr = []byte("---\nvm_extensions: null")
				})

				It("initializes an empty slice of strings", func() {
					expected := serviceadapter.VMExtensions{}

					err := yaml.Unmarshal(yamlStr, &actual)

					Expect(err).NotTo(HaveOccurred())
					Expect(actual.VMExtensions).To(Equal(expected))
				})
			})
		})
	})

	Describe("plan", func() {
		Describe("(de)serialising from/to JSON", func() {
			planJson := []byte(`{
				"instance_groups": [
					{
						"name": "example-server",
						"vm_type": "small",
						"vm_extensions": ["public_ip"],
						"persistent_disk_type": "ten",
						"networks": [
							"example-network"
						],
						"azs": [
							"example-az"
						],
						"instances": 1,
						"lifecycle": "errand"
					}
				],
				"properties": {
					"example": "property"
				},
				"update": {
					"canaries": 1,
					"max_in_flight": 10,
					"canary_watch_time": "1000-30000",
					"update_watch_time": "1000-30000",
					"serial": false
				}
			}`)

			expectedPlan := serviceadapter.Plan{
				InstanceGroups: []serviceadapter.InstanceGroup{{
					Name:               "example-server",
					VMType:             "small",
					VMExtensions:       []string{"public_ip"},
					PersistentDiskType: "ten",
					Networks:           []string{"example-network"},
					AZs:                []string{"example-az"},
					Instances:          1,
					Lifecycle:          "errand",
				}},
				Properties: serviceadapter.Properties{"example": "property"},
				Update: &serviceadapter.Update{
					Canaries:        1,
					MaxInFlight:     10,
					CanaryWatchTime: "1000-30000",
					UpdateWatchTime: "1000-30000",
					Serial:          booleanPointer(false),
				},
			}

			It("deserialises plan object containing all optional fields from json", func() {
				var plan serviceadapter.Plan
				Expect(json.Unmarshal(planJson, &plan)).To(Succeed())
				Expect(plan).To(Equal(expectedPlan))
			})

			It("serialises plan object containing all optional fields to json", func() {
				Expect(toJson(expectedPlan)).To(MatchJSON(planJson))
			})

			It("serialises plan object containing only mandatory fields to json", func() {
				expectedPlan := serviceadapter.Plan{
					InstanceGroups: []serviceadapter.InstanceGroup{{
						Name:      "example-server",
						VMType:    "small",
						Networks:  []string{"example-network"},
						AZs:       []string{"az1"},
						Instances: 1,
					}},
					Properties: serviceadapter.Properties{},
				}

				planJson := []byte(`{
					"instance_groups": [
						{
							"name": "example-server",
							"vm_type": "small",
							"networks": [
								"example-network"
							],
							"instances": 1,
							"azs": ["az1"]
						}
					],
					"properties": {}
				}`)
				Expect(toJson(expectedPlan)).To(MatchJSON(planJson))
			})

			It("serializes required fields for instance_groups", func() {
				plan := serviceadapter.Plan{
					InstanceGroups: []serviceadapter.InstanceGroup{{}},
					Properties:     serviceadapter.Properties{},
				}

				planJson := toJson(plan)

				Expect(planJson).To(ContainSubstring("name"))
				Expect(planJson).To(ContainSubstring("vm_type"))
				Expect(planJson).To(ContainSubstring("instances"))
				Expect(planJson).To(ContainSubstring("networks"))
				Expect(planJson).To(ContainSubstring("azs"))
			})

			It("serialises required fields for update", func() {
				plan := serviceadapter.Plan{
					InstanceGroups: []serviceadapter.InstanceGroup{{}},
					Properties:     serviceadapter.Properties{},
					Update:         &serviceadapter.Update{},
				}

				planJson := toJson(plan)

				Expect(planJson).To(ContainSubstring("canaries"))
				Expect(planJson).To(ContainSubstring("max_in_flight"))
				Expect(planJson).To(ContainSubstring("canary_watch_time"))
				Expect(planJson).To(ContainSubstring("update_watch_time"))
			})

			It("does not serialise optional fields for update", func() {
				plan := serviceadapter.Plan{
					InstanceGroups: []serviceadapter.InstanceGroup{{}},
					Properties:     serviceadapter.Properties{},
					Update:         &serviceadapter.Update{},
				}

				planJson := toJson(plan)

				Expect(planJson).NotTo(ContainSubstring("serial"))
			})
		})

		Describe("validation", func() {
			var plan serviceadapter.Plan

			BeforeEach(func() {
				plan = serviceadapter.Plan{
					InstanceGroups: []serviceadapter.InstanceGroup{{
						Name:      "example-server",
						VMType:    "small",
						Networks:  []string{"example-network"},
						Instances: 1,
						AZs:       []string{"az1"},
					}},
					Properties: serviceadapter.Properties{},
				}
			})

			Context("when nothing is missing", func() {
				It("returns no error", func() {
					Expect(plan.Validate()).ToNot(HaveOccurred())
				})
			})

			Context("when instance groups are missing", func() {
				BeforeEach(func() {
					plan.InstanceGroups = nil
				})

				It("returns an error", func() {
					Expect(plan.Validate()).To(HaveOccurred())
				})
			})

			Context("when vm type is missing", func() {
				BeforeEach(func() {
					plan.InstanceGroups[0].VMType = ""
				})

				It("returns an error", func() {
					Expect(plan.Validate()).To(HaveOccurred())
				})
			})

			Context("when networks is missing", func() {
				BeforeEach(func() {
					plan.InstanceGroups[0].Networks = nil
				})

				It("returns an error", func() {
					Expect(plan.Validate()).To(HaveOccurred())
				})
			})

			Context("when azs are missing", func() {
				BeforeEach(func() {
					plan.InstanceGroups[0].AZs = nil
				})

				It("returns an error", func() {
					Expect(plan.Validate()).To(HaveOccurred())
				})
			})

			Context("when azs are empty", func() {
				BeforeEach(func() {
					plan.InstanceGroups[0].AZs = []string{}
				})

				It("returns an error", func() {
					Expect(plan.Validate()).To(HaveOccurred())
				})
			})

			Context("when instances is 0", func() {
				BeforeEach(func() {
					plan.InstanceGroups[0].Instances = 0
				})

				It("returns an error", func() {
					Expect(plan.Validate()).To(HaveOccurred())
				})
			})

			Context("when instance group name is missing", func() {
				BeforeEach(func() {
					plan.InstanceGroups[0].Name = ""
				})

				It("returns an error", func() {
					Expect(plan.Validate()).To(HaveOccurred())
				})
			})
		})
	})
})
