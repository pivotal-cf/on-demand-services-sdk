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

package serviceadapter_test

import (
	"github.com/pivotal-cf/on-demand-services-sdk/bosh"
	. "github.com/pivotal-cf/on-demand-services-sdk/serviceadapter"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Instance Groups Mapping", func() {
	var (
		stemcell                = "windows-ME"
		deploymentGroupsAndJobs = map[string][]string{
			"real-instance-group":    []string{"important-job", "extra-job"},
			"another-instance-group": []string{"underrated-job"},
		}

		instanceGroups  []InstanceGroup
		serviceReleases ServiceReleases

		manifestInstanceGroups []bosh.InstanceGroup
		generateErr            error
	)

	BeforeEach(func() {
		instanceGroups = []InstanceGroup{
			{
				Name:               "real-instance-group",
				VMType:             "a-vm",
				VMExtensions:       []string{"what an extension"},
				PersistentDiskType: "such-persistence",
				Instances:          7,
				Networks:           []string{"an-etwork", "another-etwork"},
				AZs:                []string{"an-az", "jay-z"},
			},
			{
				Name:               "another-instance-group",
				VMType:             "another-vm",
				PersistentDiskType: "such-persistence",
				Instances:          7,
				Networks:           []string{"another-etwork"},
				AZs:                []string{"another-az"},
			},
		}

		serviceReleases = ServiceReleases{
			{Name: "real-release", Version: "4", Jobs: []string{"important-job"}},
			{Name: "good-release", Version: "doesn't matter", Jobs: []string{"extra-job"}},
			{Name: "service-backups", Version: "doesn't matter", Jobs: []string{"underrated-job"}},
		}
	})

	JustBeforeEach(func() {
		manifestInstanceGroups, generateErr = GenerateInstanceGroupsWithNoProperties(instanceGroups, serviceReleases, stemcell, deploymentGroupsAndJobs)
	})

	Context("when each instance group and job is provided", func() {
		It("generates deployment instance groups", func() {
			Expect(manifestInstanceGroups).To(ConsistOf(bosh.InstanceGroup{
				Name:               "real-instance-group",
				Instances:          7,
				VMType:             "a-vm",
				VMExtensions:       []string{"what an extension"},
				PersistentDiskType: "such-persistence",
				Networks:           []bosh.Network{{Name: "an-etwork"}, {Name: "another-etwork"}},
				AZs:                []string{"an-az", "jay-z"},
				Stemcell:           stemcell,
				Jobs: []bosh.Job{
					{Name: "important-job", Release: "real-release"},
					{Name: "extra-job", Release: "good-release"},
				},
			},
				bosh.InstanceGroup{
					Name:               "another-instance-group",
					Instances:          7,
					VMType:             "another-vm",
					PersistentDiskType: "such-persistence",
					Networks:           []bosh.Network{{Name: "another-etwork"}},
					AZs:                []string{"another-az"},
					Stemcell:           stemcell,
					Jobs: []bosh.Job{
						{Name: "underrated-job", Release: "service-backups"},
					},
				},
			))
		})

		It("returns no error", func() {
			Expect(generateErr).NotTo(HaveOccurred())
		})
	})

	Context("when no instance groups are provided", func() {
		BeforeEach(func() {
			instanceGroups = nil
		})

		It("returns an error", func() {
			Expect(generateErr).To(MatchError(MatchRegexp(`^no instance groups provided$`)))
		})
	})

	Context("when providing an instance group that's not expected", func() {
		BeforeEach(func() {
			instanceGroups = append(instanceGroups, InstanceGroup{Name: "i am not wanted"})
		})

		It("returns no error", func() {
			Expect(generateErr).NotTo(HaveOccurred())
		})

		It("does not include the unexpected instance group", func() {
			for _, manifestInstanceGroup := range manifestInstanceGroups {
				Expect(manifestInstanceGroup.Name).NotTo(Equal("i am not wanted"))
			}
		})
	})

	Context("when a job is expected but not provided", func() {
		BeforeEach(func() {
			serviceReleases[1].Jobs = nil
		})

		It("returns an error", func() {
			Expect(generateErr).To(MatchError("job 'extra-job' not provided"))
		})
	})

	Context("when an expected job is provided twice", func() {
		BeforeEach(func() {
			serviceReleases = append(serviceReleases, ServiceRelease{Name: "doppelganger", Jobs: []string{"underrated-job"}})
		})

		It("returns an error", func() {
			Expect(generateErr).To(MatchError(ContainSubstring("job 'underrated-job' provided 2 times")))
		})
	})
})
