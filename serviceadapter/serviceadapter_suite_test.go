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

package serviceadapter_test

import (
	"errors"
	"fmt"
	"io"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"
	"github.com/pivotal-cf/on-demand-services-sdk/bosh"
	"github.com/pivotal-cf/on-demand-services-sdk/serviceadapter"

	"encoding/json"
	"testing"

	"gopkg.in/yaml.v2"
)

func TestServiceadapter(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Service adapter Suite")
}

func toJson(obj interface{}) string {
	str, err := json.Marshal(obj)
	Expect(err).NotTo(HaveOccurred())
	return string(str)
}

func toYaml(obj interface{}) string {
	str, err := yaml.Marshal(obj)
	Expect(err).NotTo(HaveOccurred())
	return string(str)
}

func defaultServiceDeployment() serviceadapter.ServiceDeployment {
	return serviceadapter.ServiceDeployment{
		DeploymentName: "service-instance-deployment",
		Releases: serviceadapter.ServiceReleases{
			{
				Name:    "release-name",
				Version: "release-version",
				Jobs:    []string{"job_one", "job_two"},
			},
		},
		Stemcells: []serviceadapter.Stemcell{{
			OS:      "BeOS",
			Version: "2",
		}},
	}
}

func defaultRequestParams() serviceadapter.RequestParameters {
	return serviceadapter.RequestParameters{"key": "foo", "bar": "baz"}
}

func defaultSecretParams() serviceadapter.ManifestSecrets {
	return serviceadapter.ManifestSecrets{"((/a/secret/path))": "some r34||y s3cr3t v41", "((another))": "one"}
}

func defaultDNSParams() serviceadapter.DNSAddresses {
	return serviceadapter.DNSAddresses{"foo": "a.b.c", "bar": "d.e.f"}
}

func defaultPlan() serviceadapter.Plan {
	return serviceadapter.Plan{
		InstanceGroups: []serviceadapter.InstanceGroup{{
			Name:               "another-example-server",
			VMType:             "small",
			PersistentDiskType: "ten",
			Networks:           []string{"example-network"},
			AZs:                []string{"example-az"},
			Instances:          1,
			Lifecycle:          "errand",
		}},
		Properties: serviceadapter.Properties{"example": "property"},
	}
}

func defaultPreviousPlan() serviceadapter.Plan {
	return serviceadapter.Plan{
		InstanceGroups: []serviceadapter.InstanceGroup{{
			Name:               "an-example-server",
			VMType:             "medium",
			PersistentDiskType: "ten",
			Networks:           []string{"example-network"},
			AZs:                []string{"example-az"},
			Instances:          1,
			Lifecycle:          "errand",
		}},
		Properties: serviceadapter.Properties{"example": "property"},
	}
}

func defaultManifest() bosh.BoshManifest {
	return bosh.BoshManifest{Name: "another-deployment-name",
		Releases: []bosh.Release{
			{
				Name:    "a-release",
				Version: "latest",
			},
		},
		InstanceGroups: []bosh.InstanceGroup{},
		Stemcells: []bosh.Stemcell{
			{
				Alias:   "greatest",
				OS:      "Windows",
				Version: "3.1",
			},
		},
	}
}

func defaultPreviousManifest() bosh.BoshManifest {
	return bosh.BoshManifest{Name: "another-deployment-name",
		Releases: []bosh.Release{
			{
				Name:    "a-release",
				Version: "latest",
			},
		},
		InstanceGroups: []bosh.InstanceGroup{},
		Stemcells: []bosh.Stemcell{
			{
				Alias:   "greatest",
				OS:      "Windows",
				Version: "3.1",
			},
		},
	}
}

func defaultPreviousBoshConfigs() serviceadapter.BOSHConfigs {
	return serviceadapter.BOSHConfigs{
		"cloud-config":   "fake-cloud-config",
		"cpi-config":     "fake-cpi-config",
		"runtime-config": "fake-runtime-config",
	}
}

type CLIErrorMatcher struct {
	exitCode       int
	errorSubstring string
}

func BeACLIError(exitCode int, errorSubstring string) types.GomegaMatcher {
	return &CLIErrorMatcher{
		exitCode:       exitCode,
		errorSubstring: errorSubstring,
	}
}

func (c CLIErrorMatcher) Match(actual interface{}) (bool, error) {
	if actual == nil {
		return false, errors.New("Expected error, none occured")
	}

	theError, ok := actual.(serviceadapter.CLIHandlerError)
	if !ok {
		return false, fmt.Errorf("Expected error to be of type serviceadapter.CLIHandlerError, instead got '%v'", actual)
	}

	if theError.ExitCode != c.exitCode {
		return false, nil
	}
	if !strings.Contains(theError.Error(), c.errorSubstring) {
		return false, nil
	}
	return true, nil
}

func (c CLIErrorMatcher) FailureMessage(actual interface{}) string {
	theError, _ := actual.(serviceadapter.CLIHandlerError)
	if theError.ExitCode != c.exitCode {
		return fmt.Sprintf("Expected Exit Code\n\t%d\nto equal\n\t%d", theError.ExitCode, c.exitCode)
	}
	return fmt.Sprintf("Expected error message\n\t\"%s\"\nto contain\n\t\"%s\"", theError.Error(), c.errorSubstring)
}

func (c CLIErrorMatcher) NegatedFailureMessage(actual interface{}) string {
	theError, _ := actual.(serviceadapter.CLIHandlerError)
	if theError.ExitCode == c.exitCode {
		return fmt.Sprintf("Expected Exit Code\n\t%d\nto not equal\n\t%d", theError.ExitCode, c.exitCode)
	}
	return fmt.Sprintf("Expected error message\n\t\"%s\"\nto not contain\n\t\"%s\"", theError.Error(), c.errorSubstring)
}

type FakeWriter struct{}

func (f *FakeWriter) Write(b []byte) (int, error) {
	return 0, errors.New("boom!")
}

func NewFakeReader() io.Reader {
	return &FakeReader{}
}

type FakeReader struct {
}

func (f *FakeReader) Read(b []byte) (int, error) {
	return 1, fmt.Errorf("fool!")
}
