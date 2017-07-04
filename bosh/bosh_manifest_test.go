// Copyright (C) 2016-Present Pivotal Software, Inc. All rights reserved.

// This program and the accompanying materials are made available under
// the terms of the under the Apache License, Version 2.0 (the "License”);
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at

// http://www.apache.org/licenses/LICENSE-2.0

// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package bosh_test

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/pivotal-cf/on-demand-services-sdk/bosh"

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
				VMExtensions:       []string{"extended"},
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
		Variables: []bosh.Variable{
			bosh.Variable{
				Name: "admin_password",
				Type: "password",
			},
			bosh.Variable{
				Name: "default_ca",
				Type: "certificate",
				Options: map[string]interface{}{
					"is_ca":             true,
					"common_name":       "some-ca",
					"alternative_names": []string{"some-other-ca"},
				},
			},
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
			Variables: []bosh.Variable{},
		}

		content, err := yaml.Marshal(emptyManifest)
		Expect(err).NotTo(HaveOccurred())

		Expect(content).NotTo(ContainSubstring("static_ips:"))
		Expect(content).NotTo(ContainSubstring("lifecycle:"))
		Expect(content).NotTo(ContainSubstring("azs:"))
		Expect(content).NotTo(ContainSubstring("vm_extensions:"))
		Expect(content).NotTo(ContainSubstring("persistent_disk_type:"))
		Expect(content).NotTo(ContainSubstring("jobs:"))
		Expect(content).NotTo(ContainSubstring("provides:"))
		Expect(content).NotTo(ContainSubstring("consumes:"))
		Expect(content).NotTo(ContainSubstring("properties:"))
		Expect(content).NotTo(ContainSubstring("serial:"))
		Expect(content).NotTo(ContainSubstring("variables:"))
	})

	It("omits optional options from Variables", func() {
		emptyManifest := bosh.BoshManifest{
			Variables: []bosh.Variable{
				bosh.Variable{
					Name: "admin_password",
					Type: "password",
				},
			},
		}

		content, err := yaml.Marshal(emptyManifest)
		Expect(err).NotTo(HaveOccurred())
		Expect(content).NotTo(ContainSubstring("options:"))
	})

})
