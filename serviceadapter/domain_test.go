package serviceadapter_test

import (
	. "github.com/pivotal-cf/on-demand-service-broker-sdk/serviceadapter"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Domain", func() {
	Context("RequestParameters", func() {
		Context("when arbitraryParams are present", func() {
			It("can extract arbitraryParams", func() {
				params := RequestParameters{"parameters": map[string]interface{}{"foo": "bar"}}
				Expect(params.ArbitraryParams()).To(Equal(map[string]interface{}{"foo": "bar"}))
			})
		})
		Context("when arbitraryParams are absent", func() {
			It("arbitrary params are empty", func() {
				params := RequestParameters{"plan_id": "baz"}
				Expect(params.ArbitraryParams()).To(Equal(map[string]interface{}{}))
			})
		})
	})
})
