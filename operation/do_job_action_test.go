package operation_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/on-demand-services-sdk/bosh"
	"github.com/pivotal-cf/on-demand-services-sdk/operation"
)

var _ = Describe("DoJobAction", func() {
	Context("when the job is present in the manifest", func() {
		It("performs the job action", func() {
			manifest := &bosh.BoshManifest{
				InstanceGroups: []bosh.InstanceGroup{
					{
						Name: "some-instance-group-1",
						Jobs: []bosh.Job{{
							Name: "gemfire-locator",
						}},
					},
					{
						Name: "some-instance-group-2",
						Jobs: []bosh.Job{{
							Name: "gemfire-locator",
						}},
					},
				},
			}

			operation.New(manifest).
				FindJob("gemfire-locator").
				DoJobAction(func(i int, job *bosh.Job) {
					job.Provides = map[string]bosh.ProvidesLink{
						"gemfire-locator-address": {Shared: true},
					}
				})

			Expect(manifest.InstanceGroups[0].Jobs[0].Provides).To(Equal(map[string]bosh.ProvidesLink{
				"gemfire-locator-address": {Shared: true},
			}))
			Expect(manifest.InstanceGroups[1].Jobs[0].Provides).To(Equal(map[string]bosh.ProvidesLink{
				"gemfire-locator-address": {Shared: true},
			}))
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
				DoJobAction(func(i int, job *bosh.Job) {
					Fail("this step should not be reached")
				}).
				Error()

			Expect(err).To(MatchError("failed to find job 'some-incorrect-job-name' within manifest"))
		})
	})
})
