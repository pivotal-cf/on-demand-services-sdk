package operation_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/on-demand-services-sdk/bosh"
	"github.com/pivotal-cf/on-demand-services-sdk/operation"
)

var _ = Describe("FindJob", func() {
	Context("when the job is present in the manifest", func() {
		It("loads the job and returns no error", func() {
			manifest := &bosh.BoshManifest{
				InstanceGroups: []bosh.InstanceGroup{{
					Jobs: []bosh.Job{{
						Name: "gemfire-locator",
					}},
				}},
			}

			err := operation.New(manifest).
				FindJob("gemfire-locator").
				Error()

			Expect(err).To(BeNil())
		})
	})

	Context("when the job is not present in the manifest", func() {
		It("returns a helpful error message", func() {
			manifest := &bosh.BoshManifest{
				InstanceGroups: []bosh.InstanceGroup{{
					Jobs: []bosh.Job{{
						Name: "some-incorrect-bosh-job",
					}},
				}},
			}

			err := operation.New(manifest).
				FindJob("gemfire-locator").
				Error()

			Expect(err).To(MatchError("failed to find job 'gemfire-locator' within manifest"))
		})
	})

	Context("when the operation already has an error", func() {
		It("returns the error message and performs no other steps", func() {
			manifest := &bosh.BoshManifest{
				InstanceGroups: []bosh.InstanceGroup{{
					Jobs: []bosh.Job{{
						Name: "some-incorrect-bosh-job",
					}},
				}},
			}

			err := operation.New(manifest).
				FindJob("gemfire-locator").
				FindJob("gemfire-server").
				FindJob("gemfire-healthcheck").
				Error()

			Expect(err).To(MatchError("failed to find job 'gemfire-locator' within manifest"))
		})
	})

	Context("when FindJob is called successively", func() {
		It("subsequent operation commands only affect latest 'Found' job", func() {
			manifest := &bosh.BoshManifest{
				InstanceGroups: []bosh.InstanceGroup{{
					Jobs: []bosh.Job{
						{Name: "gemfire-locator"},
						{Name: "gemfire-server"},
					},
				}},
			}

			operation.New(manifest).
				FindJob("gemfire-locator").
				FindJob("gemfire-server").
				DoJobAction(func(i int, job *bosh.Job) {
					job.Name = "some-changed-job-name"
				})

			Expect(manifest.InstanceGroups[0].Jobs[0].Name).To(Equal("gemfire-locator"))
			Expect(manifest.InstanceGroups[0].Jobs[1].Name).To(Equal("some-changed-job-name"))
		})
	})

})
