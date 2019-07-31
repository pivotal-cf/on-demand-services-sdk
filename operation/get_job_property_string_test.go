package operation_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/on-demand-services-sdk/bosh"
	"github.com/pivotal-cf/on-demand-services-sdk/operation"
)

var _ = Describe("GetJobPropertyString", func() {
	Context("when the job is present in the manifest and nested paths are specified", func() {
		It("fetches the property as a string", func() {
			manifest := &bosh.BoshManifest{
				InstanceGroups: []bosh.InstanceGroup{{
					Jobs: []bosh.Job{{
						Name: "gemfire-locator",
						Properties: map[string]interface{}{
							"gemfire": map[interface{}]interface{}{
								"tls": map[interface{}]interface{}{
									"enabled": "true",
								},
							},
						},
					}},
				}},
			}

			result, err := operation.New(manifest).
				FindJob("gemfire-locator").
				GetJobPropertyString("gemfire/tls/enabled")
			Expect(err).NotTo(HaveOccurred())

			Expect(result).To(Equal("true"))
		})
	})

	Context("when the job is present in the manifest and path query is specified", func() {
		It("fetches the property as a string", func() {
			manifest := &bosh.BoshManifest{
				InstanceGroups: []bosh.InstanceGroup{{
					Jobs: []bosh.Job{{
						Name: "route_registrar",
						Properties: map[string]interface{}{
							"route_registrar": map[interface{}]interface{}{
								"routes": []interface{}{
									map[interface{}]interface{}{
										"name": "cloudcache",
										"port": 8080,
										"registration_interval": "20s",
									},
								},
							},
						},
					}},
				}},
			}

			result, err := operation.New(manifest).
				FindJob("route_registrar").
				GetJobPropertyString("route_registrar/routes/name=cloudcache/registration_interval")
			Expect(err).NotTo(HaveOccurred())

			Expect(result).To(Equal("20s"))
		})
	})

	Context("when the job is present in the manifest and a trailing path index is specified", func() {
		It("fetches the property as a string", func() {
			manifest := &bosh.BoshManifest{
				InstanceGroups: []bosh.InstanceGroup{{
					Jobs: []bosh.Job{{
						Name: "route_registrar",
						Properties: map[string]interface{}{
							"route_registrar": map[interface{}]interface{}{
								"routes": []interface{}{
									map[interface{}]interface{}{
										"name": "cloudcache",
										"port": 8080,
										"registration_interval": "20s",
										"uris": []interface{}{"some-uri-1", "some-uri-2"},
									},
								},
							},
						},
					}},
				}},
			}

			result, err := operation.New(manifest).
				FindJob("route_registrar").
				GetJobPropertyString("route_registrar/routes/name=cloudcache/uris/1")
			Expect(err).NotTo(HaveOccurred())

			Expect(result).To(Equal("some-uri-2"))
		})
	})

	Context("when the job is present in the manifest and a trailing path index is specified but there's no string", func() {
		It("returns an error message", func() {
			manifest := &bosh.BoshManifest{
				InstanceGroups: []bosh.InstanceGroup{{
					Jobs: []bosh.Job{{
						Name: "route_registrar",
						Properties: map[string]interface{}{
							"route_registrar": map[interface{}]interface{}{
								"routes": []interface{}{
									map[interface{}]interface{}{
										"name": "cloudcache",
										"port": 8080,
										"registration_interval": "20s",
										"uris": []interface{}{123, 456},
									},
								},
							},
						},
					}},
				}},
			}

			_, err := operation.New(manifest).
				FindJob("route_registrar").
				GetJobPropertyString("route_registrar/routes/name=cloudcache/uris/1")

			Expect(err).To(MatchError("failed to find string value at 'route_registrar/routes/name=cloudcache/uris/1', instead '456'(int) was found"))
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
										"name": "cloudcache",
										"port": 8080,
										"registration_interval": "20s",
										"uris": []interface{}{"some-uri-1", "some-uri-2"},
									},
								},
							},
						},
					}},
				}},
			}

			_, err := operation.New(manifest).
				FindJob("route_registrar").
				GetJobPropertyString("route_registrar/routes/name=cloudcache/uris/100")

			Expect(err).To(MatchError(MatchRegexp("failed to find value at 'route_registrar/routes/name=cloudcache/uris/100', because .* only has 2 values")))
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
										"name": "cloudcache",
										"port": 8080,
										"registration_interval": "20s",
										"uris": []interface{}{"some-uri-1", "some-uri-2"},
									},
								},
							},
						},
					}},
				}},
			}

			_, err := operation.New(manifest).
				FindJob("route_registrar").
				GetJobPropertyString("route_registrar/routes/name=cloudcache/uris/some-non-digit-key")

			Expect(err).To(MatchError(MatchRegexp("failed to find value at 'route_registrar/routes/name=cloudcache/uris/some-non-digit-key', because .* was found but a non-digit was specified at .some-non-digit-key")))
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
										"name": "cloudcache",
										"port": 8080,
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
				GetJobPropertyString("route_registrar/routes/name=some-incorrect-property/some_key")

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
				GetJobPropertyString("gemfire/tls/enabled")

			Expect(err).To(MatchError("failed to find string value at 'gemfire/tls/enabled', instead 'true'(bool) was found"))
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
									"enabled": "true",
								},
							},
						},
					}},
				}},
			}

			_, err := operation.New(manifest).
				FindJob("some-incorrect-job").
				GetJobPropertyString("gemfire/tls/enabled")

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
										"enabled": "true",
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
										"enabled": "true",
									},
								},
							},
						}},
					},
				},
			}

			_, err := operation.New(manifest).
				FindJob("gemfire-locator").
				GetJobPropertyString("gemfire/tls/enabled")

			Expect(err).To(MatchError("failed to execute 'GetJobPropertyString': not implemented for cases where multiple jobs are retrieved"))
		})
	})
})
