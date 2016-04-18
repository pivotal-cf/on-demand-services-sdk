package serviceadapter_test

import (
	"encoding/json"

	"github.com/pivotal-cf/on-demand-service-broker-sdk/serviceadapter"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("BoshInfo", func() {
	Describe("(De)serialising JSON", func() {
		boshInfoJSON := []byte(`{
	  "name": "a-name",
	  "stemcell_os": "BeOS",
	  "stemcell_version": "2"
	}`)

		expectedBoshInfo := serviceadapter.BoshInfo{
			Name:            "a-name",
			StemcellOS:      "BeOS",
			StemcellVersion: "2",
		}

		It("deserialises a BoshInfo object from JSON", func() {
			var boshInfo serviceadapter.BoshInfo
			Expect(json.Unmarshal(boshInfoJSON, &boshInfo)).To(Succeed())
			Expect(boshInfo).To(Equal(expectedBoshInfo))
		})

		It("serialises a BoshInfo object to JSON", func() {
			Expect(json.Marshal(expectedBoshInfo)).To(MatchJSON(boshInfoJSON))
		})
	})

	Describe("validation", func() {
		It("returns no error when all fields non-empty", func() {
			boshInfo := serviceadapter.BoshInfo{
				Name:            "a-name",
				StemcellOS:      "BeOS",
				StemcellVersion: "2",
			}
			Expect(boshInfo.Validate()).To(Succeed())
		})

		It("returns an error when a fields is empty", func() {
			boshInfo := serviceadapter.BoshInfo{
				Name:       "a-name",
				StemcellOS: "BeOS",
				// StemcellVersion: "2",
			}
			Expect(boshInfo.Validate()).NotTo(Succeed())
		})
	})
})
