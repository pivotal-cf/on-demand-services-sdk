package bosh_test

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/pivotal-cf/on-demand-service-broker-sdk/bosh"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"gopkg.in/yaml.v2"
)

var _ = Describe("de(serialising) BOSH manifests", func() {
	boolPointer := func(b bool) *bool {
		return &b
	}

	sampleManifest := bosh.BoshManifest{
		Name: "deployment-name",
		Releases: []bosh.Release{
			{
				Name:    "a-release",
				Version: "latest",
			},
		},
		Stemcells: []bosh.Stemcell{
			{
				Alias:   "greatest",
				OS:      "Windows",
				Version: "3.1",
			},
		},
		InstanceGroups: []bosh.InstanceGroup{
			{
				Name:      "jerb",
				Instances: 1,
				Jobs: []bosh.Job{
					{
						Name:    "broker",
						Release: "a-release",
						Provides: map[string]bosh.ProvidesLink{
							"some_link": {As: "link-name"},
						},
						Consumes: map[string]interface{}{
							"another_link":   bosh.ConsumesLink{From: "jerb-link"},
							"nullified_link": "nil",
						},
						Properties: map[string]interface{}{
							"some_property": "some_value",
						},
					},
				},
				VMType:             "massive",
				PersistentDiskType: "big",
				AZs:                []string{"az1", "az2"},
				Stemcell:           "greatest",
				Networks: []bosh.Network{
					{
						Name:      "a-network",
						StaticIPs: []string{"10.0.0.0"},
						Default:   []string{"dns"},
					},
				},
			},
			{
				Name:      "an-errand",
				Lifecycle: "errand",
				Instances: 1,
				Jobs: []bosh.Job{
					{
						Name:    "a-release",
						Release: "a-release",
					},
				},
				VMType:   "small",
				Stemcell: "greatest",
				Networks: []bosh.Network{
					{
						Name: "a-network",
					},
				},
			},
		},
		Properties: map[string]interface{}{
			"foo": "bar",
		},
		Update: bosh.Update{
			Canaries:        1,
			CanaryWatchTime: "30000-180000",
			UpdateWatchTime: "30000-180000",
			MaxInFlight:     4,
			Serial:          boolPointer(false),
		},
	}

	It("serialises bosh manifests", func() {
		cwd, err := os.Getwd()
		Expect(err).NotTo(HaveOccurred())
		manifestBytes, err := ioutil.ReadFile(filepath.Join(cwd, "fixtures", "manifest.yml"))
		Expect(err).NotTo(HaveOccurred())

		serialisedManifest, err := yaml.Marshal(sampleManifest)
		Expect(err).NotTo(HaveOccurred())
		Expect(serialisedManifest).To(MatchYAML(manifestBytes))
	})

	It("omits optional keys", func() {
		emptyManifest := bosh.BoshManifest{
			Releases: []bosh.Release{
				{},
			},
			Stemcells: []bosh.Stemcell{
				{},
			},
			InstanceGroups: []bosh.InstanceGroup{
				{
					Networks: []bosh.Network{
						{},
					},
				},
			},
			Update: bosh.Update{
				Canaries:        1,
				CanaryWatchTime: "30000-180000",
				UpdateWatchTime: "30000-180000",
				MaxInFlight:     4,
			},
		}

		content, err := yaml.Marshal(emptyManifest)
		Expect(err).NotTo(HaveOccurred())

		Expect(content).NotTo(ContainSubstring("static_ips:"))
		Expect(content).NotTo(ContainSubstring("lifecycle:"))
		Expect(content).NotTo(ContainSubstring("azs:"))
		Expect(content).NotTo(ContainSubstring("persistent_disk_type:"))
		Expect(content).NotTo(ContainSubstring("jobs:"))
		Expect(content).NotTo(ContainSubstring("provides:"))
		Expect(content).NotTo(ContainSubstring("consumes:"))
		Expect(content).NotTo(ContainSubstring("properties:"))
		Expect(content).NotTo(ContainSubstring("serial:"))
	})
})
