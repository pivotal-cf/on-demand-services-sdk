// Copyright (C) 2016-Present Pivotal Software, Inc. All rights reserved.

// This program and the accompanying materials are made available under
// the terms of the under the Apache License, Version 2.0 (the "License");
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
	"errors"
	"fmt"

	yaml "gopkg.in/yaml.v2"

	"github.com/pivotal-cf/brokerapi"
	"github.com/pivotal-cf/on-demand-services-sdk/serviceadapter"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/on-demand-services-sdk/bosh"
)

var _ = Describe("Domain", func() {
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

		Context("when bindResource is present", func() {
			It("can extract bindResource", func() {
				params := serviceadapter.RequestParameters{"bind_resource": map[string]interface{}{"app_guid": "foo"}}
				Expect(params.BindResource()).To(Equal(brokerapi.BindResource{AppGuid: "foo"}))
			})
		})

		Context("when bindResource is absent", func() {
			It("bindResources is empty", func() {
				params := serviceadapter.RequestParameters{"plan_id": "baz"}
				Expect(params.BindResource()).To(Equal(brokerapi.BindResource{}))
			})
		})

		Context("when bindResource is not a JSON", func() {
			It("bindResources is empty", func() {
				params := serviceadapter.RequestParameters{"bind_resource": func() {}}
				Expect(params.BindResource()).To(Equal(brokerapi.BindResource{}))
			})
		})

		Context("Arbitrary Context", func() {
			It("is empty when not passed", func() {
				params := serviceadapter.RequestParameters{"plan_id": "baz"}
				Expect(params.ArbitraryContext()).To(Equal(map[string]interface{}{}))
			})

			It("extracts the context", func() {
				expectedContext := map[string]interface{}{
					"platform":   "cloudfoundry",
					"space_guid": "final",
				}
				params := serviceadapter.RequestParameters{
					"context": expectedContext,
				}
				Expect(params.ArbitraryContext()).To(Equal(expectedContext))
			})
		})

		Context("Platform", func() {
			It("is empty when not passed", func() {
				params := serviceadapter.RequestParameters{"plan_id": "baz"}
				Expect(params.Platform()).To(BeEmpty())
			})

			It("extracts the platform from context if present", func() {
				expectedContext := map[string]interface{}{
					"platform":   "cloudfoundry",
					"space_guid": "final",
				}
				params := serviceadapter.RequestParameters{
					"context": expectedContext,
				}
				Expect(params.Platform()).To(Equal("cloudfoundry"))
			})

			It("is empty if context exists but it has no platform", func() {
				expectedContext := map[string]interface{}{
					"space_guid": "final",
				}
				params := serviceadapter.RequestParameters{
					"context": expectedContext,
				}
				Expect(params.Platform()).To(BeEmpty())
			})

			It("is empty if platform is not a string", func() {
				expectedContext := map[string]interface{}{
					"platform":   1,
					"space_guid": "final",
				}
				params := serviceadapter.RequestParameters{
					"context": expectedContext,
				}
				Expect(params.Platform()).To(BeEmpty())
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
					"properties": {},
					"lifecycle_errands": {
					}
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

			It("fails with an error when deserialising a JSON without update.max_in_flight property", func() {
				j := []byte(
					`{
						"lifecycle_errands": {
							"post_deploy": {
								"name": "health-check",
								"instances": ["redis-server/0"]
							},
							"pre_delete": {
								"name": "cleanup-data",
								"instances": ["redis-server/0"]
							}
						},
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
								"lifecycle": "errand",
								"migrated_from": [
									{"name": "old-server"}
								]
							}
						],
						"properties": {
							"example": "property"
						},
						"update": {
							"canaries": 1,
							"canary_watch_time": "1000-30000",
							"update_watch_time": "1000-30000",
							"serial": false
						}	
					}`)

				var p serviceadapter.Plan
				Expect(json.Unmarshal(j, &p)).To(MatchError("MaxInFlight must be either an integer or a percentage. Got <nil>"))
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

		DescribeTable(
			"unmarshalling from JSON",
			func(maxInFlight bosh.MaxInFlightValue, expectedErr error) {
				j := jsonPlanWithMaxInFlight(maxInFlight)

				var plan serviceadapter.Plan
				err := json.Unmarshal(j, &plan)

				if expectedErr != nil {
					Expect(err).To(MatchError(expectedErr))
				} else {
					Expect(err).NotTo(HaveOccurred())
					Expect(plan).To(Equal(planWithMaxInFlight(maxInFlight)))
				}
			},
			Entry("a percentage", "25%", nil),
			Entry("an integer", 4, nil),
			Entry("a float", 0.2, errors.New("MaxInFlight must be either an integer or a percentage. Got 0.2")),
			Entry("a bool", true, errors.New("MaxInFlight must be either an integer or a percentage. Got true")),
			Entry("a non percentage string", "some instances", errors.New("MaxInFlight must be either an integer or a percentage. Got some instances")),
		)

		DescribeTable(
			"marshalling to JSON",
			func(maxInFlight bosh.MaxInFlightValue, expectedErr error) {
				p := planWithMaxInFlight(maxInFlight)

				j, err := json.Marshal(&p)
				if expectedErr != nil {
					Expect(err).To(HaveOccurred())
					e, ok := err.(*json.MarshalerError)
					Expect(ok).To(BeTrue())
					Expect(e.Err).To(MatchError(expectedErr))
				} else {
					Expect(err).NotTo(HaveOccurred())
					Expect(j).To(MatchJSON(jsonPlanWithMaxInFlight(maxInFlight)))
				}
			},
			Entry("a percentage", "25%", nil),
			Entry("an integer", 4, nil),
			Entry("a float", 0.2, errors.New("MaxInFlight must be either an integer or a percentage. Got 0.2")),
			Entry("a bool", true, errors.New("MaxInFlight must be either an integer or a percentage. Got true")),
			Entry("a non percentage string", "some instances", errors.New("MaxInFlight must be either an integer or a percentage. Got some instances")),
		)

		DescribeTable(
			"unmarshalling from YAML",
			func(maxInFlight bosh.MaxInFlightValue, expectedErr error) {
				y := yamlPlanWithMaxInFlight(maxInFlight)

				var p serviceadapter.Plan
				err := yaml.Unmarshal(y, &p)
				if expectedErr != nil {
					Expect(err).To(MatchError(expectedErr))
				} else {
					Expect(err).NotTo(HaveOccurred())
					Expect(p).To(Equal(planWithMaxInFlight(maxInFlight)))
				}
			},
			Entry("a percentage", "25%", nil),
			Entry("an integer", 4, nil),
			Entry("a float", 0.2, errors.New("MaxInFlight must be either an integer or a percentage. Got 0.2")),
			Entry("null", "null", errors.New("MaxInFlight must be either an integer or a percentage. Got <nil>")),
			Entry("null", "~", errors.New("MaxInFlight must be either an integer or a percentage. Got <nil>")),
			Entry("a bool", true, errors.New("MaxInFlight must be either an integer or a percentage. Got true")),
			Entry("a non percentage string", "some instances", errors.New("MaxInFlight must be either an integer or a percentage. Got some instances")),
		)

		DescribeTable(
			"marshalling to YAML",
			func(maxInFlight bosh.MaxInFlightValue, expectedErr error) {
				p := planWithMaxInFlight(maxInFlight)

				content, err := yaml.Marshal(p)
				if expectedErr != nil {
					Expect(err).To(MatchError(expectedErr))
				} else {
					Expect(err).NotTo(HaveOccurred())
					Expect(content).To(MatchYAML(yamlPlanWithMaxInFlight(maxInFlight)))
				}
			},
			Entry("a percentage", "25%", nil),
			Entry("an integer", 4, nil),
			Entry("a float", 0.2, errors.New("MaxInFlight must be either an integer or a percentage. Got 0.2")),
			Entry("nil", nil, errors.New("MaxInFlight must be either an integer or a percentage. Got <nil>")),
			Entry("a bool", true, errors.New("MaxInFlight must be either an integer or a percentage. Got true")),
			Entry("a non percentage string", "some instances", errors.New("MaxInFlight must be either an integer or a percentage. Got some instances")),
		)
	})
})

func planWithMaxInFlight(maxInFlight bosh.MaxInFlightValue) serviceadapter.Plan {
	return serviceadapter.Plan{
		LifecycleErrands: serviceadapter.LifecycleErrands{
			PostDeploy: []serviceadapter.Errand{{
				Name:      "health-check",
				Instances: []string{"redis-server/0"},
			}},
			PreDelete: []serviceadapter.Errand{{
				Name:      "cleanup-data",
				Instances: []string{"redis-server/0"},
			}},
		},
		InstanceGroups: []serviceadapter.InstanceGroup{{
			Name:               "example-server",
			VMType:             "small",
			VMExtensions:       []string{"public_ip"},
			PersistentDiskType: "ten",
			Networks:           []string{"example-network"},
			AZs:                []string{"example-az"},
			Instances:          1,
			Lifecycle:          "errand",
			MigratedFrom: []serviceadapter.Migration{
				{Name: "old-server"},
			},
		}},
		Properties: serviceadapter.Properties{"example": "property"},
		Update: &serviceadapter.Update{
			Canaries:        1,
			MaxInFlight:     maxInFlight,
			CanaryWatchTime: "1000-30000",
			UpdateWatchTime: "1000-30000",
			Serial:          booleanPointer(false),
		},
	}
}

func jsonPlanWithMaxInFlight(maxInFlight bosh.MaxInFlightValue) []byte {
	var m bosh.MaxInFlightValue

	switch maxInFlight.(type) {
	case string:
		m = fmt.Sprintf("\"%s\"", maxInFlight)
	default:
		m = maxInFlight
	}

	return []byte(fmt.Sprintf(`{
		"lifecycle_errands": {
			"post_deploy": [{
				"name": "health-check",
				"instances": ["redis-server/0"]
			}],
			"pre_delete": [{
				"name": "cleanup-data",
				"instances": ["redis-server/0"]
			}]
		},
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
				"lifecycle": "errand",
				"migrated_from": [
					{"name": "old-server"}
				]
			}
		],
		"properties": {
			"example": "property"
		},
		"update": {
			"canaries": 1,
			"max_in_flight": %v,
			"canary_watch_time": "1000-30000",
			"update_watch_time": "1000-30000",
			"serial": false
		}
	}`, m))
}

func yamlPlanWithMaxInFlight(maxInFlight bosh.MaxInFlightValue) []byte {
	return []byte(fmt.Sprintf(`
---
lifecycle_errands:
  post_deploy:
   - name: health-check
     instances:
     - redis-server/0
  pre_delete:
    - name: cleanup-data
      instances:
      - redis-server/0
instance_groups:
- name: example-server
  vm_type: small
  vm_extensions:
  - public_ip
  persistent_disk_type: ten
  networks:
  - example-network
  azs:
  - example-az
  instances: 1
  lifecycle: errand
  migrated_from:
  - name: old-server
properties:
  example: property
update:
  canaries: 1
  max_in_flight: %v
  canary_watch_time: 1000-30000
  update_watch_time: 1000-30000
  serial: false
`, maxInFlight))
}

func booleanPointer(b bool) *bool {
	return &b
}
