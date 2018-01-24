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
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"text/template"

	"github.com/pivotal-cf/on-demand-services-sdk/bosh"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"gopkg.in/yaml.v2"
)

var _ = Describe("(de)serialising BOSH manifests", func() {
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
				MigratedFrom: []bosh.Migration{
					{
						Name: "old-instance-group-name",
					},
				},
				Env: map[string]interface{}{
					"bosh": map[string]interface{}{
						"password":                "passwerd",
						"keep_root_password":      true,
						"remove_dev_tools":        false,
						"remove_static_libraries": false,
						"swap_size":               0,
					},
					"something_else": "foo",
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
		Update: &bosh.Update{
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
		Tags: map[string]interface{}{
			"quadrata":  "parrot",
			"secondTag": "tagValue",
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
			Update: &bosh.Update{
				Canaries:        1,
				CanaryWatchTime: "30000-180000",
				UpdateWatchTime: "30000-180000",
				MaxInFlight:     4,
			},
			Variables: []bosh.Variable{},
			Tags:      map[string]interface{}{},
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
		Expect(content).NotTo(ContainSubstring("migrated_from:"))
		Expect(content).NotTo(ContainSubstring("tags:"))
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

	DescribeTable(
		"marshalling when max in flight set to",
		func(maxInFlight bosh.MaxInFlightValue, expectedErr error, expectedContent string) {
			manifest := bosh.BoshManifest{
				Update: &bosh.Update{
					MaxInFlight: maxInFlight,
				},
			}
			content, err := yaml.Marshal(&manifest)

			if expectedErr != nil {
				Expect(err).To(MatchError(expectedErr))
			} else {
				Expect(err).NotTo(HaveOccurred())
				Expect(string(content)).To(ContainSubstring(expectedContent))
			}
		},
		Entry("a percentage", "25%", nil, "max_in_flight: 25%"),
		Entry("an integer", 4, nil, "max_in_flight: 4"),
		Entry("a float", 0.2, errors.New("MaxInFlight must be either an integer or a percentage. Got 0.2"), ""),
		Entry("nil", nil, errors.New("MaxInFlight must be either an integer or a percentage. Got <nil>"), ""),
		Entry("a bool", true, errors.New("MaxInFlight must be either an integer or a percentage. Got true"), ""),
		Entry("a non percentage string", "some instances", errors.New("MaxInFlight must be either an integer or a percentage. Got some instances"), ""),
		Entry("a numeric string", "24", errors.New("MaxInFlight must be either an integer or a percentage. Got 24"), ""),
	)

	DescribeTable(
		"unmarshalling when max in flight set to",
		func(maxInFlight bosh.MaxInFlightValue, expectedErr error) {
			cwd, err := os.Getwd()
			Expect(err).NotTo(HaveOccurred())
			tmpl, err := template.ParseFiles(filepath.Join(cwd, "fixtures", "manifest_template.yml"))
			Expect(err).NotTo(HaveOccurred())

			type params struct {
				MaxInFlight bosh.MaxInFlightValue
			}
			p := params{maxInFlight}

			output := gbytes.NewBuffer()
			err = tmpl.Execute(output, p)
			Expect(err).NotTo(HaveOccurred())

			var manifest bosh.BoshManifest
			err = yaml.Unmarshal(output.Contents(), &manifest)

			if expectedErr != nil {
				Expect(err).To(MatchError(expectedErr))
			} else {
				Expect(err).NotTo(HaveOccurred())
				Expect(manifest.Update.MaxInFlight).To(Equal(maxInFlight))
			}
		},
		Entry("a percentage", "25%", nil),
		Entry("an integer", 4, nil),
		Entry("a float", 0.2, errors.New("MaxInFlight must be either an integer or a percentage. Got 0.2")),
		Entry("null", "null", errors.New("MaxInFlight must be either an integer or a percentage. Got <nil>")),
		Entry("a bool", true, errors.New("MaxInFlight must be either an integer or a percentage. Got true")),
		Entry("a non percentage string", "some instances", errors.New("MaxInFlight must be either an integer or a percentage. Got some instances")),
	)
})
