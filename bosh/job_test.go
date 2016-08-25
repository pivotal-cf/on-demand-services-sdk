package bosh_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/on-demand-service-broker-sdk/bosh"
)

var _ = Describe("bosh jobs", func() {
	It("can add links", func() {
		job := bosh.Job{}.
			AddConsumesLink("foo", "a-job").
			AddConsumesLink("bar", "other-job")
		Expect(job.Consumes["foo"]).To(Equal(bosh.ConsumesLink{From: "a-job"}))
		Expect(job.Consumes["bar"]).To(Equal(bosh.ConsumesLink{From: "other-job"}))
	})

	It("can add nullified links", func() {
		job := bosh.Job{}.AddNullifiedConsumesLink("not-wired")
		Expect(job.Consumes["not-wired"]).To(Equal("nil")) // Yes, this really should be string "nil"
	})
})
