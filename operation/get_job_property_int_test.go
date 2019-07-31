package operation_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/on-demand-services-sdk/bosh"
	"github.com/pivotal-cf/on-demand-services-sdk/operation"
)

var _ = Describe("GetJobPropertyInt", func() {
	Context("when the job is present in the manifest and nested paths are specified", func() {
		It("fetches the property as an int", func() {
			manifest := &bosh.BoshManifest{
				InstanceGroups: []bosh.InstanceGroup{{
					Jobs: []bosh.Job{{
						Name: "gemfire-locator",
						Properties: map[string]interface{}{
							"gemfire": map[interface{}]interface{}{
								"tls": map[interface{}]interface{}{
									"enabled": 123,
								},
							},
						},
					}},
				}},
			}

			result, err := operation.New(manifest).
				FindJob("gemfire-locator").
				GetJobPropertyInt("gemfire/tls/enabled")
			Expect(err).NotTo(HaveOccurred())

			Expect(result).To(Equal(123))
		})
	})

	Context("when the job is present in the manifest and path query is specified", func() {
		It("fetches the property as an int", func() {
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
										"registration_interval": 20,
									},
								},
							},
						},
					}},
				}},
			}

			result, err := operation.New(manifest).
				FindJob("route_registrar").
				GetJobPropertyInt("route_registrar/routes/name=cloudcache/registration_interval")
			Expect(err).NotTo(HaveOccurred())

			Expect(result).To(Equal(20))
		})
	})

	Context("when the job is present in the manifest and a trailing path index is specified", func() {
		It("fetches the property as an int", func() {
			manifest := &bosh.BoshManifest{
				InstanceGroups: []bosh.InstanceGroup{{
					Jobs: []bosh.Job{{
						Name: "route_registrar",
						Properties: map[string]interface{}{
							"route_registrar": map[interface{}]interface{}{
								"routes": []interface{}{
									map[interface{}]interface{}{
										"name":                   "cloudcache",
										"port":                   8080,
										"registration_intervals": []interface{}{10, 20},
									},
								},
							},
						},
					}},
				}},
			}

			result, err := operation.New(manifest).
				FindJob("route_registrar").
				GetJobPropertyInt("route_registrar/routes/name=cloudcache/registration_intervals/1")
			Expect(err).NotTo(HaveOccurred())

			Expect(result).To(Equal(20))
		})
	})

	Context("when the job is present in the manifest and a trailing path index is specified but there's no int", func() {
		It("returns an error message", func() {
			manifest := &bosh.BoshManifest{
				InstanceGroups: []bosh.InstanceGroup{{
					Jobs: []bosh.Job{{
						Name: "route_registrar",
						Properties: map[string]interface{}{
							"route_registrar": map[interface{}]interface{}{
								"routes": []interface{}{
									map[interface{}]interface{}{
										"name":                   "cloudcache",
										"port":                   8080,
										"registration_intervals": []interface{}{"some-value-1", "some-value-2"},
									},
								},
							},
						},
					}},
				}},
			}

			_, err := operation.New(manifest).
				FindJob("route_registrar").
				GetJobPropertyInt("route_registrar/routes/name=cloudcache/registration_intervals/1")

			Expect(err).To(MatchError("failed to find int value at 'route_registrar/routes/name=cloudcache/registration_intervals/1', instead 'some-value-2'(string) was found"))
		})
	})

	Context("when the job is present in the manifest and there is no trailing path for array", func() {
		It("returns an error message", func() {
			manifest := &bosh.BoshManifest{
				InstanceGroups: []bosh.InstanceGroup{{
					Jobs: []bosh.Job{{
						Name: "route_registrar",
						Properties: map[string]interface{}{
							"route_registrar": map[interface{}]interface{}{
								"routes": []interface{}{
									map[interface{}]interface{}{
										"name":                   "cloudcache",
										"port":                   8080,
										"registration_intervals": []interface{}{10, 20},
									},
								},
							},
						},
					}},
				}},
			}

			_, err := operation.New(manifest).
				FindJob("route_registrar").
				GetJobPropertyInt("route_registrar/routes/name=cloudcache/registration_intervals/100")

			Expect(err).To(MatchError(MatchRegexp("failed to find value at 'route_registrar/routes/name=cloudcache/registration_intervals/100', because .* only has 2 values")))
		})
	})

	Context("when the job is present in the manifest and there is no trailing path for array", func() {
		It("returns an error message", func() {
			manifest := &bosh.BoshManifest{
				InstanceGroups: []bosh.InstanceGroup{{
					Jobs: []bosh.Job{{
						Name: "route_registrar",
						Properties: map[string]interface{}{
							"route_registrar": map[interface{}]interface{}{
								"routes": []interface{}{
									map[interface{}]interface{}{
										"name":                   "cloudcache",
										"port":                   8080,
										"registration_intervals": []interface{}{10, 20},
									},
								},
							},
						},
					}},
				}},
			}

			_, err := operation.New(manifest).
				FindJob("route_registrar").
				GetJobPropertyInt("route_registrar/routes/name=cloudcache/registration_intervals/some-non-digit-key")

			Expect(err).To(MatchError(MatchRegexp("failed to find value at 'route_registrar/routes/name=cloudcache/registration_intervals/some-non-digit-key', because .* was found but a non-digit was specified at .some-non-digit-key")))
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

			_, err := operation.New(manifest).
				FindJob("route_registrar").
				GetJobPropertyInt("route_registrar/routes/name=some-incorrect-property/some_key")

			Expect(err).To(MatchError(ContainSubstring("failed match 'name=some-incorrect-property' of 'route_registrar/routes/name=some-incorrect-property/some_key' in:")))
		})
	})

	Context("when the job is present in the manifest and path query is specified but the type is not a string", func() {
		It("returns a helpful error message", func() {
			manifest := &bosh.BoshManifest{
				InstanceGroups: []bosh.InstanceGroup{{
					Jobs: []bosh.Job{{
						Name: "gemfire-locator",
						Properties: map[string]interface{}{
							"gemfire": map[interface{}]interface{}{
								"tls": map[interface{}]interface{}{
									"enabled": true,
								},
							},
						},
					}},
				}},
			}

			_, err := operation.New(manifest).
				FindJob("gemfire-locator").
				GetJobPropertyInt("gemfire/tls/enabled")

			Expect(err).To(MatchError("failed to find int value at 'gemfire/tls/enabled', instead 'true'(bool) was found"))
		})
	})

	Context("when the job is not present in the manifest", func() {
		It("returns a helpful error message", func() {
			manifest := &bosh.BoshManifest{
				InstanceGroups: []bosh.InstanceGroup{{
					Jobs: []bosh.Job{{
						Name: "gemfire-locator",
						Properties: map[string]interface{}{
							"gemfire": map[interface{}]interface{}{
								"tls": map[interface{}]interface{}{
									"enabled": 123,
								},
							},
						},
					}},
				}},
			}

			_, err := operation.New(manifest).
				FindJob("some-incorrect-job").
				GetJobPropertyInt("gemfire/tls/enabled")

			Expect(err).To(MatchError("failed to find job 'some-incorrect-job' within manifest"))
		})
	})

	Context("when multiple jobs are found for the given operation", func() {
		It("returns a helpful error message", func() {
			manifest := &bosh.BoshManifest{
				InstanceGroups: []bosh.InstanceGroup{
					{
						Name: "some-instance-group-1",
						Jobs: []bosh.Job{{
							Name: "gemfire-locator",
							Properties: map[string]interface{}{
								"gemfire": map[interface{}]interface{}{
									"tls": map[interface{}]interface{}{
										"enabled": 123,
									},
								},
							},
						}},
					},
					{
						Name: "some-instance-group-2",
						Jobs: []bosh.Job{{
							Name: "gemfire-locator",
							Properties: map[string]interface{}{
								"gemfire": map[interface{}]interface{}{
									"tls": map[interface{}]interface{}{
										"enabled": 123,
									},
								},
							},
						}},
					},
				},
			}

			_, err := operation.New(manifest).
				FindJob("gemfire-locator").
				GetJobPropertyInt("gemfire/tls/enabled")

			Expect(err).To(MatchError("failed to execute 'GetJobPropertyInt': not implemented for cases where multiple jobs are retrieved"))
		})
	})
})
