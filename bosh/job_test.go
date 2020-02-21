// Copyright (C) 2016-Present Pivotal Software, Inc. All rights reserved.

// This program and the accompanying materials are made available under
// the terms of the under the Apache License, Version 2.0 (the "License");
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
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/on-demand-services-sdk/bosh"
)

var _ = Describe("bosh jobs", func() {
	It("can add consumer links", func() {
		job := bosh.Job{}.
			AddConsumesLink("foo", "a-job").
			AddConsumesLink("bar", "other-job")
		Expect(job.Consumes["foo"]).To(Equal(bosh.ConsumesLink{From: "a-job"}))
		Expect(job.Consumes["bar"]).To(Equal(bosh.ConsumesLink{From: "other-job"}))
	})

	It("can add provider links", func() {
		job := bosh.Job{}.
			AddSharedProvidesLink("foo").
			AddSharedProvidesLink("bar")
		Expect(job.Provides["foo"]).To(Equal(bosh.ProvidesLink{Shared: true}))
		Expect(job.Provides["bar"]).To(Equal(bosh.ProvidesLink{Shared: true}))
	})

	It("can add provider links with aliases", func() {
		alias := bosh.Alias{
			Domain:             "some.domain.internal",
			HealthFilter:       "healthy",
			InitialHealthCheck: "asynchronous",
		}

		job := bosh.Job{}.AddProvidesLinkWithAliases("mylink", []bosh.Alias{alias})

		Expect(job.Provides["mylink"].Aliases).To(Equal([]bosh.Alias{alias}))
	})

	It("can cross deployment links", func() {
		job := bosh.Job{}.AddCrossDeploymentConsumesLink("foo", "a-job", "a-deployment")
		Expect(job.Consumes["foo"]).To(Equal(bosh.ConsumesLink{From: "a-job", Deployment: "a-deployment"}))
	})

	It("can add nullified links", func() {
		job := bosh.Job{}.AddNullifiedConsumesLink("not-wired")
		Expect(job.Consumes["not-wired"]).To(Equal("nil")) // Yes, this really should be string "nil"
	})

	It("can add custom provider definitions", func() {
		job := bosh.Job{}.AddCustomProviderDefinition("some-name", "some-type", []string{"prop1"})
		job = job.AddCustomProviderDefinition("some-other", "some-other-type", nil)
		Expect(job.CustomProviderDefinitions).To(
			ConsistOf(
				bosh.CustomProviderDefinition{Name: "some-name", Type: "some-type", Properties: []string{"prop1"}},
				bosh.CustomProviderDefinition{Name: "some-other", Type: "some-other-type"},
			),
		)
	})
})
