package operation_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/on-demand-services-sdk/bosh"
	"github.com/pivotal-cf/on-demand-services-sdk/operation"
)

var _ = Describe("EachInstanceGroup", func() {
	Context("By default", func() {
		It("performs an action on each instance group", func() {
			manifest := &bosh.BoshManifest{
				InstanceGroups: []bosh.InstanceGroup{
					{Name: "some-instance-group-1"},
					{Name: "some-instance-group-2"},
				},
			}

			operation.New(manifest).
				EachInstanceGroup(func(ig *bosh.InstanceGroup) {
					ig.Name = "some-modified-instance-group"
				})

			Expect(manifest.InstanceGroups[0].Name).To(Equal("some-modified-instance-group"))
			Expect(manifest.InstanceGroups[1].Name).To(Equal("some-modified-instance-group"))
		})
	})

	Context("when the operation already has an error", func() {
		It("returns the error message and performs no other steps", func() {
			manifest := &bosh.BoshManifest{
				InstanceGroups: []bosh.InstanceGroup{
					{Name: "some-instance-group-1"},
					{Name: "some-instance-group-2"},
				},
			}

			err := operation.New(manifest).
				FindJob("some-incorrect-job-name").
				EachInstanceGroup(func(ig *bosh.InstanceGroup) {
					ig.Name = "some-modified-instance-group"
				}).
				Error()

			Expect(err).To(MatchError("failed to find job 'some-incorrect-job-name' within manifest"))
			Expect(manifest.InstanceGroups[0].Name).To(Equal("some-instance-group-1"))
			Expect(manifest.InstanceGroups[1].Name).To(Equal("some-instance-group-2"))
		})
	})
})
