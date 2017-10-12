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
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"

	"gopkg.in/yaml.v2"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	"github.com/pivotal-cf/on-demand-services-sdk/bosh"
	"github.com/pivotal-cf/on-demand-services-sdk/serviceadapter"
	"github.com/pivotal-cf/on-demand-services-sdk/serviceadapter/testharness/testvariables"
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

	doNotImplementInterfaces = false

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
		stdout = new(bytes.Buffer)
		stderr = new(bytes.Buffer)

		resetTempFiles()
	})

	Describe("command without arguments", func() {
		It("logs and exits with 1", func() {
			exitCode = startPassingCommandAndGetExitCode([]string{})

			Expect(exitCode).To(Equal(1))
			Expect(stderr.String()).To(Equal("[odb-sdk] the following commands are supported: generate-manifest, create-binding, delete-binding, dashboard-url\n"))
		})
	})

	Describe("an unknown subcommand argument", func() {
		It("logs and exits with 1", func() {
			exitCode = startPassingCommandAndGetExitCode([]string{"non-existing-subcommand"})

			Expect(exitCode).To(Equal(1))
			Expect(stderr.String()).To(ContainSubstring(`[odb-sdk] unknown subcommand: non-existing-subcommand. The following commands are supported: generate-manifest, create-binding, delete-binding, dashboard-url`))
		})
	})

	Describe("generate-manifest without optional parameters", func() {
		It("succeeds", func() {
			exitCode = startPassingCommandAndGetExitCode([]string{"generate-manifest",
				toJson(expectedServiceDeployment),
				toJson(expectedCurrentPlan),
				toJson(expectedRequestParams),
				"",
				"null",
			})
			actualServiceDeployment, actualCurrentPlan, actualRequestParams, actualPreviousManifest, actualPreviousPlan := decodeAndVerifyGenerateManifestResponse()

			Expect(exitCode).To(Equal(0))
			Expect(actualServiceDeployment).To(Equal(expectedServiceDeployment))
			Expect(actualCurrentPlan).To(Equal(expectedCurrentPlan))
			Expect(actualRequestParams).To(Equal(expectedRequestParams))
			Expect(actualPreviousManifest).To(BeNil())
			Expect(actualPreviousPlan).To(BeNil())
			Expect(stdout.String()).To(Equal(toYaml(expectedResultantBoshManifest)))
		})
	})

	Describe("generate-manifest with optional parameters", func() {
		It("succeeds", func() {
			exitCode = startPassingCommandAndGetExitCode([]string{
				"generate-manifest",
				toJson(expectedServiceDeployment),
				toJson(expectedCurrentPlan),
				toJson(expectedRequestParams),
				toYaml(expectedPreviousManifest),
				toJson(expectedPreviousPlan),
			})
			_, _, _, actualPreviousManifest, actualPreviousPlan := decodeAndVerifyGenerateManifestResponse()

			Expect(actualPreviousManifest).To(Equal(&expectedPreviousManifest))
			Expect(actualPreviousPlan).To(Equal(&expectedPreviousPlan))
		})
	})

	Describe("generate-manifest with missing arguments error", func() {
		It("logs and exits with 1", func() {
			exitCode = startPassingCommandAndGetExitCode([]string{"generate-manifest"})

			Expect(exitCode).To(Equal(1))
			Expect(stderr.String()).To(ContainSubstring(
				"Missing arguments for generate-manifest. Usage: testharness generate-manifest <service-deployment-JSON> <plan-JSON> <request-params-JSON> <previous-manifest-YAML> <previous-plan-JSON>"))
		})
	})

	Describe("generate-manifest with general error", func() {
		It("exits 10 and logs", func() {
			exitCode = startFailingCommandAndGetExitCode([]string{"generate-manifest",
				toJson(expectedServiceDeployment),
				toJson(expectedCurrentPlan),
				toJson(expectedRequestParams),
				"",
				"null",
			}, "true")
			decodeAndVerifyGenerateManifestResponse()

			Expect(exitCode).To(Equal(1))
			Expect(stdout.String()).To(Equal("some message to the user"))
		})
	})

	Describe("create-binding", func() {
		It("succeeds", func() {
			exitCode = startPassingCommandAndGetExitCode([]string{"create-binding", expectedBindingID, toJson(expectedBoshVMs), toYaml(expectedManifest), toJson(expectedRequestParams)})
			actualBindingId, actualBoshVMs, actualBoshManifest, actualBindingParams := decodeAndVerifyCreateBindingResponse()

			Expect(exitCode).To(Equal(0))
			Expect(actualBindingId).To(Equal(expectedBindingID))
			Expect(actualBoshVMs).To(Equal(expectedBoshVMs))
			Expect(actualBoshManifest).To(Equal(expectedManifest))
			Expect(actualBindingParams).To(Equal(expectedRequestParams))
			Expect(stdout.String()).To(MatchJSON(toJson(testvariables.SuccessfulBinding)))
		})
	})

	Describe("create-binding with missing arguments error", func() {
		It("logs and exits with 1", func() {
			exitCode = startPassingCommandAndGetExitCode([]string{"create-binding"})

			Expect(exitCode).To(Equal(1))
			Expect(stderr.String()).To(ContainSubstring(
				"Missing arguments for create-binding. Usage: testharness create-binding <binding-ID> <bosh-VMs-JSON> <manifest-YAML> <request-params-JSON>",
			))
		})
	})

	Describe("create-binding where binding already exists", func() {
		It("exits with 49", func() {
			exitCode = startFailingCommandAndGetExitCode([]string{"create-binding", expectedBindingID, toJson(expectedBoshVMs), toYaml(expectedManifest), toJson(expectedRequestParams)}, testvariables.ErrBindingAlreadyExists)
			decodeAndVerifyCreateBindingResponse()

			Expect(exitCode).To(Equal(49))
		})
	})

	Describe("create-binding where app_guid isn't provided", func() {
		It("exits with 42", func() {
			exitCode = startFailingCommandAndGetExitCode([]string{"create-binding", expectedBindingID, toJson(expectedBoshVMs), toYaml(expectedManifest), toJson(expectedRequestParams)}, testvariables.ErrAppGuidNotProvided)
			decodeAndVerifyCreateBindingResponse()

			Expect(exitCode).To(Equal(42))
		})
	})

	Describe("create-binding fails where there is an internal error", func() {
		It("exits with failure", func() {
			exitCode = startFailingCommandAndGetExitCode([]string{"create-binding", expectedBindingID, toJson(expectedBoshVMs), toYaml(expectedManifest), toJson(expectedRequestParams)}, "true")
			decodeAndVerifyCreateBindingResponse()

			Expect(exitCode).To(Equal(1))
			Expect(stdout.String()).To(Equal("An internal error occured."))
		})
	})

	Describe("delete-binding", func() {
		It("succeeds", func() {
			exitCode = startPassingCommandAndGetExitCode([]string{"delete-binding", expectedBindingID, toJson(expectedBoshVMs), toYaml(expectedManifest), toJson(expectedUnbindingRequestParams)})
			actualBindingId, actualBoshVMs, actualBoshManifest, actualRequestParams := decodeAndVerifyDeleteBindingResponse()

			Expect(exitCode).To(Equal(0))
			Expect(actualBindingId).To(Equal(expectedBindingID))
			Expect(actualBoshVMs).To(Equal(expectedBoshVMs))
			Expect(actualBoshManifest).To(Equal(expectedManifest))
			Expect(actualRequestParams).To(Equal(expectedUnbindingRequestParams))
		})
	})

	Describe("delete-binding with missing arguments error", func() {
		It("logs and exits with 1", func() {
			exitCode = startPassingCommandAndGetExitCode([]string{"delete-binding"})

			Expect(exitCode).To(Equal(1))
			Expect(stderr.String()).To(ContainSubstring(
				"Missing arguments for delete-binding. Usage: testharness delete-binding <binding-ID> <bosh-VMs-JSON> <manifest-YAML> <request-params-JSON>",
			))
		})
	})

	Describe("delete-binding where binding not found", func() {
		It("exits with 41", func() {
			exitCode = startFailingCommandAndGetExitCode([]string{"delete-binding", expectedBindingID, toJson(expectedBoshVMs), toYaml(expectedManifest), toJson(expectedUnbindingRequestParams)}, testvariables.ErrBindingNotFound)
			decodeAndVerifyDeleteBindingResponse()

			Expect(exitCode).To(Equal(41))
		})
	})

	Describe("delete-binding with general error", func() {
		It("exits with a failure", func() {
			exitCode = startFailingCommandAndGetExitCode([]string{"delete-binding", expectedBindingID, toJson(expectedBoshVMs), toYaml(expectedManifest), toJson(expectedUnbindingRequestParams)}, "true")
			decodeAndVerifyDeleteBindingResponse()

			Expect(exitCode).To(Equal(1))
			Expect(stdout.String()).To(Equal("An error occurred"))
		})
	})

	Describe("dashboard-url", func() {
		It("succeeds", func() {
			exitCode = startPassingCommandAndGetExitCode([]string{"dashboard-url", "instance-identifier", toJson(expectedPlan), toYaml(expectedManifest)})
			actualPlan, actualManifest, actualInstanceID := decodeAndVerifyDashboardURLResponse()

			Expect(exitCode).To(Equal(0))
			Expect(actualInstanceID).To(Equal("instance-identifier"))
			Expect(actualPlan).To(Equal(expectedPlan))
			Expect(actualManifest).To(Equal(expectedManifest))
			Expect(stdout.Bytes()).To(MatchJSON(`{ "dashboard_url": "http://dashboard.com"}`))
		})
	})

	Describe("dashboard-url with missing arguments error", func() {
		It("logs and exits with 1", func() {
			exitCode = startPassingCommandAndGetExitCode([]string{"dashboard-url"})

			Expect(exitCode).To(Equal(1))
			Expect(stderr.String()).To(ContainSubstring(
				"Missing arguments for dashboard-url. Usage: testharness dashboard-url <instance-ID> <plan-JSON> <manifest-YAML>",
			))
		})
	})

	Describe("dashboard-url with general error", func() {
		It("exits with failure", func() {
			exitCode = startFailingCommandAndGetExitCode([]string{"dashboard-url", "instance-identifier", toJson(expectedPlan), toYaml(expectedManifest)}, "true")
			decodeAndVerifyDashboardURLResponse()

			Expect(exitCode).To(Equal(1))
			Expect(stdout.String()).To(Equal("An error occurred"))
		})
	})

	Describe("no interfaces have been implemented", func() {
		It("unknown subcommand exits with 1", func() {
			exitCode = startEmptyImplementationCommandAndGetExitCode([]string{"non-existing-subcommand"})

			Expect(exitCode).To(Equal(1))
			Expect(stderr.String()).To(ContainSubstring(`[odb-sdk] unknown subcommand: non-existing-subcommand. The following commands are supported:`))
		})

		It("generate-manifest exits with 10", func() {
			exitCode = startEmptyImplementationCommandAndGetExitCode([]string{"generate-manifest", toJson(expectedServiceDeployment), toJson(expectedCurrentPlan), toJson(expectedRequestParams), "", "null"})

			Expect(exitCode).To(Equal(10))
		})

		It("create-binding exits with 10", func() {
			exitCode = startEmptyImplementationCommandAndGetExitCode([]string{"create-binding", toJson(expectedServiceDeployment), toJson(expectedCurrentPlan), toJson(expectedRequestParams), "", "null"})

			Expect(exitCode).To(Equal(10))
		})

		It("delete-binding exits with 10", func() {
			exitCode = startEmptyImplementationCommandAndGetExitCode([]string{"delete-binding", toJson(expectedServiceDeployment), toJson(expectedCurrentPlan), toJson(expectedRequestParams), "", "null"})

			Expect(exitCode).To(Equal(10))
		})

		It("dashboard-url exits with 10", func() {
			exitCode = startEmptyImplementationCommandAndGetExitCode([]string{"dashboard-url", "id", toJson(expectedCurrentPlan), "null"})

			Expect(exitCode).To(Equal(10))
		})
	})
})

func decodeAndVerifyGenerateManifestResponse() (serviceadapter.ServiceDeployment, serviceadapter.Plan, map[string]interface{}, *bosh.BoshManifest, *serviceadapter.Plan) {
	var (
		actualServiceDeployment serviceadapter.ServiceDeployment
		actualCurrentPlan       serviceadapter.Plan
		actualRequestParams     map[string]interface{}
		actualPreviousManifest  *bosh.BoshManifest
		actualPreviousPlan      *serviceadapter.Plan
	)

	jsonDeserialise(serviceDeploymentFilePath, &actualServiceDeployment)
	jsonDeserialise(planFilePath, &actualCurrentPlan)
	jsonDeserialise(requestParamsFilePath, &actualRequestParams)
	yamlDeserialise(previousManifestFilePath, &actualPreviousManifest)
	jsonDeserialise(previousPlanFilePath, &actualPreviousPlan)

	return actualServiceDeployment, actualCurrentPlan, actualRequestParams, actualPreviousManifest, actualPreviousPlan
}

func decodeAndVerifyCreateBindingResponse() (string, bosh.BoshVMs, bosh.BoshManifest, map[string]interface{}) {
	var (
		actualBindingId     string
		actualBoshVMs       bosh.BoshVMs
		actualBoshManifest  bosh.BoshManifest
		actualBindingParams map[string]interface{}
	)

	jsonDeserialise(bindingIdFilePath, &actualBindingId)
	jsonDeserialise(boshVMsFilePath, &actualBoshVMs)
	yamlDeserialise(boshManifestFilePath, &actualBoshManifest)
	jsonDeserialise(bindingParamsFilePath, &actualBindingParams)

	return actualBindingId, actualBoshVMs, actualBoshManifest, actualBindingParams
}

func decodeAndVerifyDeleteBindingResponse() (string, bosh.BoshVMs, bosh.BoshManifest, serviceadapter.RequestParameters) {
	var (
		actualBindingId     string
		actualBoshVMs       bosh.BoshVMs
		actualBoshManifest  bosh.BoshManifest
		actualRequestParams serviceadapter.RequestParameters
	)

	jsonDeserialise(bindingIdFilePath, &actualBindingId)
	jsonDeserialise(boshVMsFilePath, &actualBoshVMs)
	yamlDeserialise(boshManifestFilePath, &actualBoshManifest)
	jsonDeserialise(bindingParamsFilePath, &actualRequestParams)

	return actualBindingId, actualBoshVMs, actualBoshManifest, actualRequestParams
}

func decodeAndVerifyDashboardURLResponse() (serviceadapter.Plan, bosh.BoshManifest, string) {
	var (
		actualInstanceID string
		actualPlan       serviceadapter.Plan
		actualManifest   bosh.BoshManifest
	)

	jsonDeserialise(instanceIDFilePath, &actualInstanceID)
	jsonDeserialise(dashboardPlanFilePath, &actualPlan)
	yamlDeserialise(dashboardManifestFilePath, &actualManifest)

	return actualPlan, actualManifest, actualInstanceID
}

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
		fmt.Sprintf("%s=%s", testvariables.GenerateManifestServiceDeploymentFileKey, serviceDeploymentFilePath),
		fmt.Sprintf("%s=%s", testvariables.GenerateManifestPlanFileKey, planFilePath),
		fmt.Sprintf("%s=%s", testvariables.GenerateManifestRequestParamsFileKey, requestParamsFilePath),
		fmt.Sprintf("%s=%s", testvariables.GenerateManifestPreviousManifestFileKey, previousManifestFilePath),
		fmt.Sprintf("%s=%s", testvariables.GenerateManifestPreviousPlanFileKey, previousPlanFilePath),

		fmt.Sprintf("%s=%s", testvariables.BindingIdFileKey, bindingIdFilePath),
		fmt.Sprintf("%s=%s", testvariables.BindingVmsFileKey, boshVMsFilePath),
		fmt.Sprintf("%s=%s", testvariables.BindingManifestFileKey, boshManifestFilePath),
		fmt.Sprintf("%s=%s", testvariables.BindingParamsFileKey, bindingParamsFilePath),

		fmt.Sprintf("%s=%s", testvariables.DashboardURLInstanceIDKey, instanceIDFilePath),
		fmt.Sprintf("%s=%s", testvariables.DashboardURLPlanKey, dashboardPlanFilePath),
		fmt.Sprintf("%s=%s", testvariables.DashboardURLManifestKey, dashboardManifestFilePath),

		fmt.Sprintf("%s=%t", testvariables.DoNotImplementInterfacesKey, doNotImplementInterfaces),
	}
}

func resetTempFiles() {
	for _, filePath := range []string{
		serviceDeploymentFilePath, planFilePath, requestParamsFilePath, previousManifestFilePath, previousPlanFilePath,
		bindingIdFilePath, boshVMsFilePath, boshManifestFilePath, bindingParamsFilePath,
		instanceIDFilePath, dashboardPlanFilePath, dashboardManifestFilePath,
	} {
		deleteFileIfExists(filePath)
	}

	createTempFilesForGenerateManifest()

	createTempFilesForBinding()

	createTempFilesForDashboardURL()
}

func deleteFileIfExists(filePath string) {
	if _, err := os.Stat(filePath); os.IsExist(err) {
		Expect(os.Remove(filePath)).To(Succeed())
	}
}

func createTempFile() string {
	file, err := ioutil.TempFile("", "sdk-tests")
	Expect(err).NotTo(HaveOccurred())
	defer file.Close()
	return file.Name()
}

func createTempFilesForGenerateManifest() {
	serviceDeploymentFilePath = createTempFile()
	planFilePath = createTempFile()
	requestParamsFilePath = createTempFile()
	previousManifestFilePath = createTempFile()
	previousPlanFilePath = createTempFile()
}

func createTempFilesForBinding() {
	bindingIdFilePath = createTempFile()
	boshVMsFilePath = createTempFile()
	boshManifestFilePath = createTempFile()
	bindingParamsFilePath = createTempFile()
}

func createTempFilesForDashboardURL() {
	instanceIDFilePath = createTempFile()
	dashboardPlanFilePath = createTempFile()
	dashboardManifestFilePath = createTempFile()
}

func jsonDeserialise(filePath string, pointer interface{}) {
	file, err := os.Open(filePath)
	Expect(err).NotTo(HaveOccurred())
	defer file.Close()
	Expect(json.NewDecoder(file).Decode(pointer)).To(Succeed())
}

func yamlDeserialise(filePath string, pointer interface{}) {
	fileBytes, err := ioutil.ReadFile(filePath)
	Expect(err).NotTo(HaveOccurred())
	Expect(yaml.Unmarshal(fileBytes, pointer)).To(Succeed())
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
