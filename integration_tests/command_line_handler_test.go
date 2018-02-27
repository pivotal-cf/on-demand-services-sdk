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

package integration_tests_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"

	"gopkg.in/yaml.v2"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	"github.com/pivotal-cf/on-demand-services-sdk/bosh"
	"github.com/pivotal-cf/on-demand-services-sdk/integration_tests/testharness/testvariables"
	"github.com/pivotal-cf/on-demand-services-sdk/serviceadapter"
)

var (
	stdout *bytes.Buffer
	stderr *bytes.Buffer

	operationFails string
	exitCode       int

	serviceDeploymentFilePath string
	planFilePath              string
	requestParamsFilePath     string
	previousManifestFilePath  string
	previousPlanFilePath      string

	bindingIdFilePath     string
	boshVMsFilePath       string
	boshManifestFilePath  string
	bindingParamsFilePath string

	instanceIDFilePath        string
	dashboardPlanFilePath     string
	dashboardManifestFilePath string

	doNotImplementInterfaces bool

	expectedServiceDeployment = serviceadapter.ServiceDeployment{
		DeploymentName: "service-instance-deployment",
		Releases: serviceadapter.ServiceReleases{
			{
				Name:    "release-name",
				Version: "release-version",
				Jobs:    []string{"job_one", "job_two"},
			},
		},
		Stemcell: serviceadapter.Stemcell{
			OS:      "BeOS",
			Version: "2",
		},
	}

	expectedCurrentPlan = serviceadapter.Plan{
		InstanceGroups: []serviceadapter.InstanceGroup{{
			Name:               "example-server",
			VMType:             "small",
			PersistentDiskType: "ten",
			Networks:           []string{"example-network"},
			AZs:                []string{"example-az"},
			Instances:          1,
			Lifecycle:          "errand",
		}},
		Properties: serviceadapter.Properties{"example": "property"},
	}

	expectedRequestParams = map[string]interface{}{"key": "foo", "bar": "baz"}

	expectedResultantBoshManifest = bosh.BoshManifest{Name: "deployment-name",
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
				Name: "Test",
				Properties: map[string]interface{}{
					"parseSymbols": "yes%[===",
				},
			},
		},
	}

	expectedPreviousPlan = serviceadapter.Plan{
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

	expectedPlan = expectedPreviousPlan

	expectedPreviousManifest = bosh.BoshManifest{Name: "another-deployment-name",
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
		}}

	expectedBindingID = "bindingId"

	expectedBoshVMs = bosh.BoshVMs{"kafka": []string{"a", "b"}}

	expectedManifest = expectedPreviousManifest

	expectedUnbindingRequestParams = serviceadapter.RequestParameters{"unbinding_param": "unbinding_value"}
)

var _ = Describe("Command line handler", func() {

	BeforeEach(func() {
		doNotImplementInterfaces = false
		stdout = new(bytes.Buffer)
		stderr = new(bytes.Buffer)
	})

	It("logs and exits with 1 if called without args", func() {
		exitCode = startPassingCommandAndGetExitCode([]string{})

		Expect(exitCode).To(Equal(1))
		Expect(stderr.String()).To(Equal("[odb-sdk] the following commands are supported: generate-manifest, create-binding, delete-binding, dashboard-url, generate-plan-schemas\n"))
	})

	It("logs and exits with 1 if called with a non-existing subcommand", func() {
		exitCode = startPassingCommandAndGetExitCode([]string{"non-existing-subcommand"})

		Expect(exitCode).To(Equal(1))
		Expect(stderr.String()).To(ContainSubstring(`[odb-sdk] unknown subcommand: non-existing-subcommand. The following commands are supported: generate-manifest, create-binding, delete-binding, dashboard-url, generate-plan-schemas`))
	})

	Describe("generate-manifest command", func() {
		It("succeeds without optional parameters", func() {
			exitCode = startPassingCommandAndGetExitCode([]string{"generate-manifest",
				toJson(expectedServiceDeployment),
				toJson(expectedCurrentPlan),
				toJson(expectedRequestParams),
				"",
				"null",
			})
			Expect(exitCode).To(Equal(0))
			Expect(stdout.String()).To(Equal(toYaml(expectedResultantBoshManifest)))
		})

		It("generate-manifest exits with 10", func() {
			exitCode = startEmptyImplementationCommandAndGetExitCode([]string{"generate-manifest", toJson(expectedServiceDeployment), toJson(expectedCurrentPlan), toJson(expectedRequestParams), "", "null"})
			Expect(exitCode).To(Equal(10))
		})

		It("succeeds with optional parameters", func() {
			exitCode = startPassingCommandAndGetExitCode([]string{
				"generate-manifest",
				toJson(expectedServiceDeployment),
				toJson(expectedCurrentPlan),
				toJson(expectedRequestParams),
				toYaml(expectedPreviousManifest),
				toJson(expectedPreviousPlan),
			})
			Expect(exitCode).To(Equal(0))
			Expect(stdout.String()).To(Equal(toYaml(expectedResultantBoshManifest)))
		})

		It("logs and exits with 1 when an argument is missing", func() {
			exitCode = startPassingCommandAndGetExitCode([]string{"generate-manifest"})

			Expect(exitCode).To(Equal(1))
			Expect(stderr.String()).To(ContainSubstring(
				"Missing arguments for generate-manifest. Usage: testharness generate-manifest <service-deployment-JSON> <plan-JSON> <request-params-JSON> <previous-manifest-YAML> <previous-plan-JSON>"))
		})

		It("exits 1 and logs when a generic error occurs", func() {
			exitCode = startFailingCommandAndGetExitCode([]string{"generate-manifest",
				toJson(expectedServiceDeployment),
				toJson(expectedCurrentPlan),
				toJson(expectedRequestParams),
				"",
				"null",
			}, "true")

			Expect(exitCode).To(Equal(1))
			Expect(stdout.String()).To(Equal("some message to the user"))
		})
	})

	Describe("create-binding command", func() {
		It("succeeds", func() {
			exitCode = startPassingCommandAndGetExitCode([]string{"create-binding", expectedBindingID, toJson(expectedBoshVMs), toYaml(expectedManifest), toJson(expectedRequestParams)})
			Expect(exitCode).To(Equal(0))
			Expect(stdout.String()).To(MatchJSON(toJson(testvariables.SuccessfulBinding)))
		})

		It("create-binding exits with 10", func() {
			exitCode = startEmptyImplementationCommandAndGetExitCode([]string{"create-binding", toJson(expectedServiceDeployment), toJson(expectedCurrentPlan), toJson(expectedRequestParams), "", "null"})
			Expect(exitCode).To(Equal(10))
		})

		It("logs and exits with 1 when an argument is missing", func() {
			exitCode = startPassingCommandAndGetExitCode([]string{"create-binding"})

			Expect(exitCode).To(Equal(1))
			Expect(stderr.String()).To(ContainSubstring(
				"Missing arguments for create-binding. Usage: testharness create-binding <binding-ID> <bosh-VMs-JSON> <manifest-YAML> <request-params-JSON>",
			))
		})

		It("exits with 49 when a binding already exists", func() {
			exitCode = startFailingCommandAndGetExitCode([]string{"create-binding", expectedBindingID, toJson(expectedBoshVMs), toYaml(expectedManifest), toJson(expectedRequestParams)}, testvariables.ErrBindingAlreadyExists)

			Expect(exitCode).To(Equal(49))
		})

		It("exits with 42 when app_guid is not provided", func() {
			exitCode = startFailingCommandAndGetExitCode([]string{"create-binding", expectedBindingID, toJson(expectedBoshVMs), toYaml(expectedManifest), toJson(expectedRequestParams)}, testvariables.ErrAppGuidNotProvided)

			Expect(exitCode).To(Equal(42))
		})

		It("exits with 1 when an error occurs", func() {
			exitCode = startFailingCommandAndGetExitCode([]string{"create-binding", expectedBindingID, toJson(expectedBoshVMs), toYaml(expectedManifest), toJson(expectedRequestParams)}, "true")

			Expect(exitCode).To(Equal(1))
			Expect(stdout.String()).To(Equal("An internal error occured."))
		})
	})

	Describe("delete-binding", func() {
		It("succeeds", func() {
			exitCode = startPassingCommandAndGetExitCode([]string{"delete-binding", expectedBindingID, toJson(expectedBoshVMs), toYaml(expectedManifest), toJson(expectedUnbindingRequestParams)})
			Expect(exitCode).To(Equal(0))
			Expect(stdout.String()).To(BeEmpty())
		})

		It("delete-binding exits with 10", func() {
			exitCode = startEmptyImplementationCommandAndGetExitCode([]string{"delete-binding", toJson(expectedServiceDeployment), toJson(expectedCurrentPlan), toJson(expectedRequestParams), "", "null"})
			Expect(exitCode).To(Equal(10))
		})

		It("logs and exits with 1 when an argument is missing", func() {
			exitCode = startPassingCommandAndGetExitCode([]string{"delete-binding"})

			Expect(exitCode).To(Equal(1))
			Expect(stderr.String()).To(ContainSubstring(
				"Missing arguments for delete-binding. Usage: testharness delete-binding <binding-ID> <bosh-VMs-JSON> <manifest-YAML> <request-params-JSON>",
			))
		})

		It("exits with 41 when the binding is not found", func() {
			exitCode = startFailingCommandAndGetExitCode([]string{"delete-binding", expectedBindingID, toJson(expectedBoshVMs), toYaml(expectedManifest), toJson(expectedUnbindingRequestParams)}, testvariables.ErrBindingNotFound)

			Expect(exitCode).To(Equal(41))
			Expect(stdout.String()).To(ContainSubstring("binding not found"))
		})

		It("exits with a failure when a generic error occurs", func() {
			exitCode = startFailingCommandAndGetExitCode([]string{"delete-binding", expectedBindingID, toJson(expectedBoshVMs), toYaml(expectedManifest), toJson(expectedUnbindingRequestParams)}, "true")

			Expect(exitCode).To(Equal(1))
			Expect(stdout.String()).To(Equal("An error occurred"))
		})
	})

	Describe("dashboard-url command", func() {
		It("succeeds", func() {
			exitCode = startPassingCommandAndGetExitCode([]string{"dashboard-url", "instance-identifier", toJson(expectedPlan), toYaml(expectedManifest)})
			Expect(exitCode).To(Equal(0))
			Expect(stdout.Bytes()).To(MatchJSON(`{ "dashboard_url": "http://dashboard.com"}`))
		})

		It("dashboard-url exits with 10", func() {
			exitCode = startEmptyImplementationCommandAndGetExitCode([]string{"dashboard-url", "id", toJson(expectedCurrentPlan), "null"})
			Expect(exitCode).To(Equal(10))
		})

		It("logs and exits with 1 when an argument is missing", func() {
			exitCode = startPassingCommandAndGetExitCode([]string{"dashboard-url"})

			Expect(exitCode).To(Equal(1))
			Expect(stderr.String()).To(ContainSubstring(
				"Missing arguments for dashboard-url. Usage: testharness dashboard-url <instance-ID> <plan-JSON> <manifest-YAML>",
			))
		})

		It("exits with 1 if a generic error occurs", func() {
			exitCode = startFailingCommandAndGetExitCode([]string{"dashboard-url", "instance-identifier", toJson(expectedPlan), toYaml(expectedManifest)}, "true")

			Expect(exitCode).To(Equal(1))
			Expect(stdout.String()).To(Equal("An error occurred"))
		})
	})

	Describe("generate-plan-schemas command", func() {
		It("returns 0 and output the schema for a plan", func() {
			exitCode = startPassingCommandAndGetExitCode([]string{"generate-plan-schemas", "--plan-json", toJson(expectedPlan)})
			schemas := serviceadapter.JSONSchemas{
				Parameters: map[string]interface{}{
					"$schema": "http://json-schema.org/draft-04/schema#",
					"type":    "object",
					"properties": map[string]interface{}{
						"billing-account": map[string]interface{}{
							"description": "Billing account number used to charge use of shared fake server.",
							"type":        "string",
						},
					},
				},
			}
			expectedPlanSchema := serviceadapter.PlanSchema{
				ServiceInstance: serviceadapter.ServiceInstanceSchema{
					Create: schemas,
					Update: schemas,
				},
				ServiceBinding: serviceadapter.ServiceBindingSchema{
					Create: schemas,
				},
			}
			Expect(exitCode).To(Equal(0))
			Expect(stdout.Bytes()).To(MatchJSON(toJson(expectedPlanSchema)))
		})

		It("returns 1 if an error occurred while parsing the CLI args", func() {
			exitCode = startPassingCommandAndGetExitCode([]string{"generate-plan-schemas"})

			Expect(exitCode).To(Equal(1))
			Expect(stderr.String()).To(ContainSubstring(
				"Incorrect arguments for generate-plan-schemas",
			))
		})

		It("returns 1 if an error occurred while generating the schema", func() {
			exitCode = startFailingCommandAndGetExitCode([]string{
				"generate-plan-schemas", "--plan-json", toJson(expectedPlan),
			}, "true")

			Expect(exitCode).To(Equal(1))
			Expect(stderr.String()).To(MatchRegexp(`\[odb-sdk\] An error occurred`))
		})

		It("returns 10 (not implemented) when the command is not implement", func() {
			exitCode = startEmptyImplementationCommandAndGetExitCode([]string{"generate-plan-schemas", "--plan-json", toJson(expectedCurrentPlan)})
			Expect(exitCode).To(Equal(10))
		})
	})
})

func startEmptyImplementationCommandAndGetExitCode(args []string) int {
	doNotImplementInterfaces = true
	operationFails = ""

	return startCommandAndGetExitCode(args)
}

func startFailingCommandAndGetExitCode(args []string, errMsg string) int {
	operationFails = errMsg

	return startCommandAndGetExitCode(args)
}

func startPassingCommandAndGetExitCode(args []string) int {
	operationFails = ""

	return startCommandAndGetExitCode(args)
}

func startCommandAndGetExitCode(args []string) int {
	cmd := exec.Command(adapterBin, args...)
	cmd.Env = resetCommandEnv()
	runningAdapter, err := gexec.Start(cmd, io.MultiWriter(GinkgoWriter, stdout), io.MultiWriter(GinkgoWriter, stderr))
	Expect(err).NotTo(HaveOccurred())
	Eventually(runningAdapter).Should(gexec.Exit())
	return runningAdapter.ExitCode()
}

func resetCommandEnv() []string {
	return []string{
		fmt.Sprintf("%s=%s", testvariables.OperationFailsKey, operationFails),
		fmt.Sprintf("%s=%t", testvariables.DoNotImplementInterfacesKey, doNotImplementInterfaces),
	}
}

func toYaml(obj interface{}) string {
	str, err := yaml.Marshal(obj)
	Expect(err).NotTo(HaveOccurred())
	return string(str)
}
func toJson(obj interface{}) string {
	str, err := json.Marshal(obj)
	Expect(err).NotTo(HaveOccurred())
	return string(str)
}
