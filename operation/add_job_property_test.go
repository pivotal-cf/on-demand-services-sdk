package operation_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/on-demand-services-sdk/bosh"
	"github.com/pivotal-cf/on-demand-services-sdk/operation"
)

var _ = Describe("AddJobProperty", func() {
	Context("when the job is present in the manifest", func() {
		It("add the property to the job", func() {
			manifest := &bosh.BoshManifest{
				InstanceGroups: []bosh.InstanceGroup{{
					Jobs: []bosh.Job{{
						Name: "gemfire-locator",
					}},
				}},
			}

			operation.New(manifest).
				FindJob("gemfire-locator").
				AddJobProperty("gemfire", true)

			Expect(manifest.InstanceGroups[0].Jobs[0].Properties).To(HaveKeyWithValue("gemfire", true))
		})
	})

	Context("when the job is present in the manifest and nested paths are specified", func() {
		It("add the property to the job", func() {
			manifest := &bosh.BoshManifest{
				InstanceGroups: []bosh.InstanceGroup{{
					Jobs: []bosh.Job{{
						Name: "gemfire-locator",
					}},
				}},
			}

			operation.New(manifest).
				FindJob("gemfire-locator").
				AddJobProperty("gemfire/tls/enabled", true)

			Expect(manifest.InstanceGroups[0].Jobs[0].Properties).To(Equal(map[string]interface{}{
				"gemfire": map[interface{}]interface{}{
					"tls": map[interface{}]interface{}{
						"enabled": true,
					},
				},
			}))
		})
	})

	Context("when the job is present in the manifest and path query is specified", func() {
		It("add the property to the job", func() {
			manifest := &bosh.BoshManifest{
				InstanceGroups: []bosh.InstanceGroup{{
					Jobs: []bosh.Job{{
						Name: "route_registrar",
						Properties: map[string]interface{}{
							"route_registrar": map[interface{}]interface{}{
								"routes": []interface{}{
									map[interface{}]interface{}{
										"name":                  "cloudcache",
										"port":                  8080,
										"registration_interval": "20s",
									},
								},
							},
						},
					}},
				}},
			}

			err := operation.New(manifest).
				FindJob("route_registrar").
				AddJobProperty("route_registrar/routes/name=cloudcache/some_key", "some_value").
				Error()
			Expect(err).NotTo(HaveOccurred())

			Expect(manifest.InstanceGroups[0].Jobs[0].Properties).To(Equal(map[string]interface{}{
				"route_registrar": map[interface{}]interface{}{
					"routes": []interface{}{
						map[interface{}]interface{}{
							"name":                  "cloudcache",
							"port":                  8080,
							"registration_interval": "20s",
							"some_key":              "some_value",
						},
					},
				},
			}))
		})
	})

	Context("when the job is present in the manifest and an incorrect path query is specified", func() {
		It("returns a helpful error message", func() {
			manifest := &bosh.BoshManifest{
				InstanceGroups: []bosh.InstanceGroup{{
					Jobs: []bosh.Job{{
						Name: "route_registrar",
						Properties: map[string]interface{}{
							"route_registrar": map[interface{}]interface{}{
								"routes": []interface{}{
									map[interface{}]interface{}{
										"name":                  "cloudcache",
										"port":                  8080,
										"registration_interval": "20s",
									},
								},
							},
						},
					}},
				}},
			}

			err := operation.New(manifest).
				FindJob("route_registrar").
				AddJobProperty("route_registrar/routes/name=some-incorrect-property/some_key", "some_value").
				Error()
			Expect(err).To(MatchError(ContainSubstring("failed match 'name=some-incorrect-property' of 'route_registrar/routes/name=some-incorrect-property/some_key' in:")))
		})
	})

	Context("when the job is present in the manifest and path already exists but is not a mapping", func() {
		It("fails with a helpful error message ", func() {
			manifest := &bosh.BoshManifest{
				InstanceGroups: []bosh.InstanceGroup{{
					Jobs: []bosh.Job{{
						Name: "gemfire-locator",
						Properties: map[string]interface{}{
							"gemfire": map[interface{}]interface{}{
								"tls": 12,
							},
						},
					}},
				}},
			}

			err := operation.New(manifest).
				FindJob("gemfire-locator").
				AddJobProperty("gemfire/tls/enabled", true).
				Error()

			Expect(err).To(MatchError("failed to apply property at 'gemfire/tls/enabled' because '12'(int) exists at .tls"))
		})
	})

	Context("when the operation already has an error", func() {
		It("returns the error message and performs no other steps", func() {
			manifest := &bosh.BoshManifest{
				InstanceGroups: []bosh.InstanceGroup{{
					Jobs: []bosh.Job{{
						Name: "gemfire-locator",
					}},
				}},
			}

			err := operation.New(manifest).
				FindJob("some-incorrect-job-name").
				AddJobProperty("gemfire", true).
				Error()

			Expect(err).To(MatchError("failed to find job 'some-incorrect-job-name' within manifest"))
			Expect(manifest.InstanceGroups[0].Jobs[0].Properties).To(BeEmpty())
		})
	})
})
