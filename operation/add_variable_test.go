package operation_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/on-demand-services-sdk/bosh"
	"github.com/pivotal-cf/on-demand-services-sdk/operation"
)

var _ = Describe("AddVariables", func() {
	Context("when the variables do not exist", func() {
		It("add variables to the manifest", func() {
			manifest := &bosh.BoshManifest{}

			operation.New(manifest).
				AddVariables(
					bosh.Variable{Name: "some-variable-1", Type: "password"},
					bosh.Variable{Name: "some-variable-2", Type: "certificate"},
				)

			Expect(manifest.Variables).To(Equal([]bosh.Variable{
				{Name: "some-variable-1", Type: "password"},
				{Name: "some-variable-2", Type: "certificate"},
			}))
		})
	})

	Context("when a variable already exists in the manifest", func() {
		It("returns a helpful error message", func() {
			manifest := &bosh.BoshManifest{}

			operation.New(manifest).
				AddVariables(bosh.Variable{Name: "some-variable-1", Type: "password"})

			err := operation.New(manifest).
				AddVariables(bosh.Variable{Name: "some-variable-1", Type: "certificate"}).
				Error()

			Expect(err).To(MatchError("failed to add bosh.Variable 'some-variable-1' because it is already present"))
		})
	})

	Context("when a variable already exists in the method call", func() {
		It("returns a helpful error message", func() {
			manifest := &bosh.BoshManifest{}

			err := operation.New(manifest).
				AddVariables(
					bosh.Variable{Name: "some-variable-1", Type: "password"},
					bosh.Variable{Name: "some-variable-1", Type: "certificate"},
				).
				Error()

			Expect(err).To(MatchError("failed to add bosh.Variable 'some-variable-1' because it is already present"))
		})
	})

	Context("when the operation already has an error", func() {
		It("returns the error message and performs no other steps", func() {
			manifest := &bosh.BoshManifest{}

			err := operation.New(manifest).
				FindJob("some-incorrect-job-name").
				AddVariables(bosh.Variable{Name: "some-variable-1", Type: "password"}).
				Error()

			Expect(err).To(MatchError("failed to find job 'some-incorrect-job-name' within manifest"))
			Expect(manifest.Variables).To(BeEmpty())
		})
	})
})
