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
	"io"
	"io/ioutil"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"

	"bytes"

	"github.com/pivotal-cf/on-demand-services-sdk/bosh"
	"github.com/pivotal-cf/on-demand-services-sdk/serviceadapter"
	"github.com/pivotal-cf/on-demand-services-sdk/serviceadapter/fakes"
)

var _ = Describe("CommandLineHandler", func() {
	const commandName = "arbitrary-command-name"

	var (
		fakeManifestGenerator     *fakes.FakeManifestGenerator
		fakeBinder                *fakes.FakeBinder
		fakeDashboardUrlGenerator *fakes.FakeDashboardUrlGenerator
		fakeSchemaGenerator       *fakes.FakeSchemaGenerator
		handler                   serviceadapter.CommandLineHandler

		serviceDeployment serviceadapter.ServiceDeployment
		args              serviceadapter.RequestParameters
		plan              serviceadapter.Plan
		previousPlan      serviceadapter.Plan
		previousManifest  bosh.BoshManifest

		serviceDeploymentJSON string
		argsJSON              string
		planJSON              string
		previousPlanJSON      string
		previousManifestYAML  string
		requestParams         serviceadapter.RequestParameters
		requestParamsJSON     string
		bindingID             string
		instanceID            string
		boshVMs               bosh.BoshVMs
		boshVMsJSON           string
		secrets               map[string]string
		secretsJSON           string

		expectedBinding serviceadapter.Binding

		outputBuffer *gbytes.Buffer
		errorBuffer  io.Writer
	)

	BeforeEach(func() {
		fakeManifestGenerator = new(fakes.FakeManifestGenerator)
		fakeBinder = new(fakes.FakeBinder)
		fakeDashboardUrlGenerator = new(fakes.FakeDashboardUrlGenerator)
		fakeSchemaGenerator = new(fakes.FakeSchemaGenerator)
		outputBuffer = gbytes.NewBuffer()
		errorBuffer = gbytes.NewBuffer()

		handler = serviceadapter.CommandLineHandler{
			ManifestGenerator:     fakeManifestGenerator,
			Binder:                fakeBinder,
			DashboardURLGenerator: fakeDashboardUrlGenerator,
			SchemaGenerator:       fakeSchemaGenerator,
		}

		serviceDeployment = defaultServiceDeployment()
		args = defaultRequestParams()
		plan = defaultPlan()
		previousPlan = defaultPreviousPlan()
		previousManifest = defaultPreviousManifest()

		requestParams = defaultRequestParams()
		requestParamsJSON = toJson(requestParams)
		bindingID = "my-binding-id"
		instanceID = "my-instance-id"
		boshVMs = bosh.BoshVMs{"kafka": []string{"a", "b"}}
		boshVMsJSON = toJson(boshVMs)
		secrets = map[string]string{"admin_pass": "pa55w0rd"}
		secretsJSON = toJson(secrets)
		expectedBinding = serviceadapter.Binding{
			Credentials: map[string]interface{}{
				"username": "alice",
			},
		}

		serviceDeploymentJSON = toJson(serviceDeployment)
		planJSON = toJson(plan)
		previousPlanJSON = toJson(previousPlan)
		argsJSON = toJson(args)
		previousManifestYAML = toYaml(previousManifest)
	})

	It("when no arguments passed returns error", func() {
		err := handler.Handle([]string{commandName}, outputBuffer, errorBuffer, bytes.NewBufferString(""))

		Expect(err).To(BeACLIError(1, "the following commands are supported: create-binding, dashboard-url, delete-binding, generate-manifest, generate-plan-schemas"))
	})

	It("does not output optional commands if not implemented", func() {
		handler.DashboardURLGenerator = nil
		handler.SchemaGenerator = nil
		err := handler.Handle([]string{commandName}, outputBuffer, errorBuffer, bytes.NewBufferString(""))

		Expect(err).To(BeACLIError(1, "the following commands are supported: create-binding, delete-binding, generate-manifest"))
		Expect(err.Error()).NotTo(ContainSubstring("dashboard-url"))
		Expect(err.Error()).NotTo(ContainSubstring("generate-plan-schemas"))
	})

	Describe("generate-manifest action", func() {
		It("succeeds with positional arguments", func() {
			manifest := bosh.BoshManifest{Name: "bill"}
			fakeManifestGenerator.GenerateManifestReturns(manifest, nil)

			err := handler.Handle([]string{
				commandName, "generate-manifest", serviceDeploymentJSON, planJSON, argsJSON, previousManifestYAML, previousPlanJSON,
			}, outputBuffer, errorBuffer, bytes.NewBufferString(""))

			Expect(err).NotTo(HaveOccurred())

			Expect(fakeManifestGenerator.GenerateManifestCallCount()).To(Equal(1))
			actualServiceDeployment, actualPlan, actualRequestParams, actualPreviousManifest, actualPreviousPlan :=
				fakeManifestGenerator.GenerateManifestArgsForCall(0)

			Expect(actualServiceDeployment).To(Equal(serviceDeployment))
			Expect(actualPlan).To(Equal(plan))
			Expect(actualRequestParams).To(Equal(args))
			Expect(actualPreviousManifest).To(Equal(&previousManifest))
			Expect(actualPreviousPlan).To(Equal(&previousPlan))

			Expect(outputBuffer).To(gbytes.Say("bill"))
		})

		It("succeeds with arguments from stdin", func() {
			rawInputParams := serviceadapter.InputParams{
				GenerateManifest: serviceadapter.GenerateManifestParams{
					ServiceDeployment: toJson(serviceDeployment),
					Plan:              toJson(plan),
					PreviousPlan:      toJson(previousPlan),
					RequestParameters: toJson(requestParams),
					PreviousManifest:  previousManifestYAML,
				},
			}
			fakeStdin := bytes.NewBuffer([]byte(toJson(rawInputParams)))
			err := handler.Handle([]string{commandName, "generate-manifest"}, outputBuffer, errorBuffer, fakeStdin)
			Expect(err).ToNot(HaveOccurred())

			actualServiceDeployment, actualPlan, actualRequestParams, actualPreviousManifest, actualPreviousPlan := fakeManifestGenerator.GenerateManifestArgsForCall(0)
			Expect(actualServiceDeployment).To(Equal(serviceDeployment))
			Expect(actualPlan).To(Equal(plan))
			Expect(actualRequestParams).To(Equal(requestParams))
			Expect(actualPreviousManifest).To(Equal(&previousManifest))
			Expect(actualPreviousPlan).To(Equal(&previousPlan))
		})

		It("returns a not-implemented error when there is no generate-manifest handler", func() {
			handler.ManifestGenerator = nil
			err := handler.Handle([]string{commandName, "generate-manifest"}, outputBuffer, errorBuffer, bytes.NewBufferString(""))

			Expect(err).To(BeACLIError(serviceadapter.NotImplementedExitCode, "generate-manifest not implemented"))

		})

		It("returns a missing args error when arguments are missing", func() {
			err := handler.Handle([]string{
				commandName, "generate-manifest", serviceDeploymentJSON,
			}, outputBuffer, errorBuffer, bytes.NewBufferString(""))

			Expect(err).To(BeACLIError(serviceadapter.ErrorExitCode, `Missing arguments for generate-manifest. Usage:`))
		})

		It("returns an error when parsing the arguments fails", func() {
			err := handler.Handle([]string{
				commandName, "generate-manifest", serviceDeploymentJSON, "asd", argsJSON, previousManifestYAML, previousPlanJSON,
			}, outputBuffer, errorBuffer, bytes.NewBufferString(""))

			Expect(err).To(MatchError(ContainSubstring("unmarshalling service plan")))
		})
	})

	Describe("create-binding action", func() {
		It("succeeds with positional arguments", func() {
			fakeBinder.CreateBindingReturns(expectedBinding, nil)

			err := handler.Handle([]string{
				commandName, "create-binding", bindingID, boshVMsJSON, previousManifestYAML, requestParamsJSON,
			}, outputBuffer, errorBuffer, bytes.NewBufferString(""))

			Expect(err).NotTo(HaveOccurred())

			Expect(fakeBinder.CreateBindingCallCount()).To(Equal(1))
			actualBindingId, actualBoshVMs, actualManifest, actualRequestParams, actualSecrets :=
				fakeBinder.CreateBindingArgsForCall(0)

			Expect(actualBindingId).To(Equal(bindingID))
			Expect(actualBoshVMs).To(Equal(boshVMs))
			Expect(actualManifest).To(Equal(previousManifest))
			Expect(actualRequestParams).To(Equal(requestParams))
			Expect(actualSecrets).To(Equal(serviceadapter.ManifestSecrets{}))

			Expect(outputBuffer).To(gbytes.Say(toJson(expectedBinding)))
		})

		It("succeeds with arguments from stdin", func() {
			rawInputParams := serviceadapter.InputParams{
				CreateBinding: serviceadapter.CreateBindingParams{
					RequestParameters: toJson(requestParams),
					BindingId:         bindingID,
					BoshVms:           toJson(boshVMs),
					Manifest:          toYaml(previousManifest),
				},
			}

			fakeBinder.CreateBindingReturns(expectedBinding, nil)
			fakeStdin := bytes.NewBuffer([]byte(toJson(rawInputParams)))

			err := handler.Handle([]string{commandName, "create-binding"}, outputBuffer, errorBuffer, fakeStdin)
			Expect(err).NotTo(HaveOccurred())

			Expect(fakeBinder.CreateBindingCallCount()).To(Equal(1))

			actualBindingId, actualBoshVMs, actualManifest, actualRequestParams, actualSecrets :=
				fakeBinder.CreateBindingArgsForCall(0)

			Expect(actualBindingId).To(Equal(bindingID))
			Expect(actualBoshVMs).To(Equal(boshVMs))
			Expect(actualManifest).To(Equal(previousManifest))
			Expect(actualRequestParams).To(Equal(requestParams))
			Expect(actualSecrets).To(Equal(serviceadapter.ManifestSecrets{}))

			Expect(outputBuffer).To(gbytes.Say(toJson(expectedBinding)))
		})

		It("returns a not-implemented error where there is no binder handler", func() {
			handler.Binder = nil
			err := handler.Handle([]string{
				commandName, "create-binding",
			}, outputBuffer, errorBuffer, bytes.NewBufferString(""))

			Expect(err).To(BeACLIError(serviceadapter.NotImplementedExitCode, "create-binding not implemented"))
		})

		It("returns a missing args error when request JSON is missing", func() {
			err := handler.Handle([]string{
				commandName, "create-binding", bindingID, boshVMsJSON, previousManifestYAML,
			}, outputBuffer, errorBuffer, bytes.NewBufferString(""))

			Expect(err).To(BeACLIError(serviceadapter.ErrorExitCode, `Missing arguments for create-binding. Usage:`))
		})

		It("returns an error when parsing the arguments fails", func() {
			boshVMsJSON += `aaa`
			err := handler.Handle([]string{
				commandName, "create-binding", bindingID, boshVMsJSON, previousManifestYAML, requestParamsJSON,
			}, outputBuffer, errorBuffer, bytes.NewBufferString(""))

			Expect(err).To(MatchError(ContainSubstring("unmarshalling BOSH VMs")))
		})

		It("returns an error when the binding cannot be created because of generic error", func() {
			fakeBinder.CreateBindingReturns(serviceadapter.Binding{}, errors.New("oops"))
			err := handler.Handle([]string{
				commandName, "create-binding", bindingID, boshVMsJSON, previousManifestYAML, requestParamsJSON,
			}, outputBuffer, errorBuffer, bytes.NewBufferString(""))

			Expect(err).To(BeACLIError(serviceadapter.ErrorExitCode, "oops"))
			Expect(outputBuffer).To(gbytes.Say("oops"))
		})
	})

	Describe("dashboard-url action", func() {
		It("succeeds with positional arguments", func() {
			fakeDashboardUrlGenerator.DashboardUrlReturns(serviceadapter.DashboardUrl{DashboardUrl: "http://url.example.com"}, nil)

			err := handler.Handle([]string{
				commandName, "dashboard-url", instanceID, planJSON, previousManifestYAML,
			}, outputBuffer, errorBuffer, bytes.NewBufferString(""))

			Expect(err).NotTo(HaveOccurred())

			Expect(fakeDashboardUrlGenerator.DashboardUrlCallCount()).To(Equal(1))
			actualInstanceID, actualPlanJSON, actualManifestYAML := fakeDashboardUrlGenerator.DashboardUrlArgsForCall(0)

			Expect(actualInstanceID).To(Equal(instanceID))
			Expect(actualPlanJSON).To(Equal(plan))
			Expect(actualManifestYAML).To(Equal(previousManifest))
			Expect(outputBuffer).To(gbytes.Say(`{"dashboard_url":"http://url.example.com"}`))
		})

		It("succeeds with arguments from stdin", func() {
			rawInputParams := serviceadapter.InputParams{
				DashboardUrl: serviceadapter.DashboardUrlParams{
					InstanceId: instanceID,
					Plan:       toJson(plan),
					Manifest:   previousManifestYAML,
				},
			}

			fakeDashboardUrlGenerator.DashboardUrlReturns(serviceadapter.DashboardUrl{DashboardUrl: "http://url.example.com"}, nil)

			fakeStdin := bytes.NewBufferString(toJson(rawInputParams))
			err := handler.Handle([]string{commandName, "dashboard-url"}, outputBuffer, errorBuffer, fakeStdin)
			Expect(err).NotTo(HaveOccurred())

			Expect(fakeDashboardUrlGenerator.DashboardUrlCallCount()).To(Equal(1))
			actualInstanceID, actualPlanJSON, actualManifestYAML := fakeDashboardUrlGenerator.DashboardUrlArgsForCall(0)

			Expect(actualInstanceID).To(Equal(instanceID))
			Expect(actualPlanJSON).To(Equal(plan))
			Expect(actualManifestYAML).To(Equal(previousManifest))
			Expect(outputBuffer).To(gbytes.Say(`{"dashboard_url":"http://url.example.com"}`))
		})

		It("returns a not-implemented error when there is no dashboard-url handler", func() {
			handler.DashboardURLGenerator = nil
			err := handler.Handle([]string{commandName, "dashboard-url"}, outputBuffer, errorBuffer, bytes.NewBufferString(""))

			Expect(err).To(BeACLIError(serviceadapter.NotImplementedExitCode, "dashboard-url not implemented"))
		})

		It("returns a missing args error when arguments are missing", func() {
			err := handler.Handle([]string{
				commandName, "dashboard-url", serviceDeploymentJSON,
			}, outputBuffer, errorBuffer, bytes.NewBufferString(""))

			Expect(err).To(BeACLIError(serviceadapter.ErrorExitCode, `Missing arguments for dashboard-url. Usage:`))
		})

		It("returns an error when parsing the arguments fails", func() {
			err := handler.Handle([]string{
				commandName, "dashboard-url", serviceDeploymentJSON, "plan", "manifest",
			}, outputBuffer, errorBuffer, bytes.NewBufferString(""))

			Expect(err).To(MatchError(ContainSubstring("unmarshalling service plan")))
		})
	})

	Describe("delete-binding action", func() {
		It("succeeds with positional arguments", func() {
			fakeBinder.DeleteBindingReturns(nil)

			err := handler.Handle([]string{
				commandName, "delete-binding", bindingID, boshVMsJSON, previousManifestYAML, requestParamsJSON,
			}, outputBuffer, errorBuffer, bytes.NewBufferString(""))

			Expect(err).NotTo(HaveOccurred())

			Expect(fakeBinder.DeleteBindingCallCount()).To(Equal(1))
			actualBindingId, actualBoshVMs, actualManifest, actualRequestParams :=
				fakeBinder.DeleteBindingArgsForCall(0)

			Expect(actualBindingId).To(Equal(bindingID))
			Expect(actualBoshVMs).To(Equal(boshVMs))
			Expect(actualManifest).To(Equal(previousManifest))
			Expect(actualRequestParams).To(Equal(requestParams))

		})

		It("succeeds with arguments from stdin", func() {
			rawInputParams := serviceadapter.InputParams{
				DeleteBinding: serviceadapter.DeleteBindingParams{
					RequestParameters: toJson(requestParams),
					BindingId:         bindingID,
					BoshVms:           toJson(boshVMs),
					Manifest:          toYaml(previousManifest),
				},
			}

			fakeBinder.DeleteBindingReturns(nil)

			fakeStdin := bytes.NewBuffer([]byte(toJson(rawInputParams)))
			err := handler.Handle([]string{commandName, "delete-binding"}, outputBuffer, errorBuffer, fakeStdin)
			Expect(err).NotTo(HaveOccurred())

			Expect(fakeBinder.DeleteBindingCallCount()).To(Equal(1))
			actualBindingId, actualBoshVMs, actualManifest, actualRequestParams :=
				fakeBinder.DeleteBindingArgsForCall(0)

			Expect(actualBindingId).To(Equal(bindingID))
			Expect(actualBoshVMs).To(Equal(boshVMs))
			Expect(actualManifest).To(Equal(previousManifest))
			Expect(actualRequestParams).To(Equal(requestParams))
		})

		It("returns a not-implemented error where there is no binder handler", func() {
			handler.Binder = nil
			err := handler.Handle([]string{
				commandName, "delete-binding",
			}, outputBuffer, errorBuffer, bytes.NewBufferString(""))

			Expect(err).To(BeACLIError(10, "delete-binding not implemented"))
		})

		It("returns a missing args error when request JSON is missing", func() {
			err := handler.Handle([]string{
				commandName, "delete-binding", previousManifestYAML}, outputBuffer, errorBuffer, bytes.NewBufferString(""))

			Expect(err).To(BeACLIError(1, "Missing arguments for delete-binding. Usage:"))
		})

		It("returns an error when parsing the arguments fails", func() {
			boshVMsJSON += `aaa`
			err := handler.Handle([]string{
				commandName, "delete-binding", bindingID, boshVMsJSON, previousManifestYAML, requestParamsJSON,
			}, outputBuffer, errorBuffer, bytes.NewBufferString(""))

			Expect(err).To(MatchError(ContainSubstring("unmarshalling BOSH VMs")))
		})

		It("returns an error when the binding cannot be created because of generic error", func() {
			fakeBinder.DeleteBindingReturns(errors.New("oops"))
			err := handler.Handle([]string{
				commandName, "delete-binding", bindingID, boshVMsJSON, previousManifestYAML, requestParamsJSON,
			}, outputBuffer, errorBuffer, bytes.NewBufferString(""))

			Expect(err).To(BeACLIError(1, "oops"))
			Expect(outputBuffer).To(gbytes.Say("oops"))
		})
	})

	Describe("generate-plan-schemas action", func() {
		It("succeeds with positional arguments", func() {
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
			fakeSchemaGenerator.GeneratePlanSchemaReturns(expectedPlanSchema, nil)

			err := handler.Handle([]string{
				commandName, "generate-plan-schemas", "--plan-json", planJSON,
			}, outputBuffer, errorBuffer, bytes.NewBufferString(""))

			Expect(err).NotTo(HaveOccurred())
			Expect(fakeSchemaGenerator.GeneratePlanSchemaCallCount()).To(Equal(1))

			Expect(fakeSchemaGenerator.GeneratePlanSchemaArgsForCall(0)).To(Equal(plan))

			contents, err := ioutil.ReadAll(outputBuffer)
			Expect(err).NotTo(HaveOccurred())
			Expect(contents).To(MatchJSON(toJson(expectedPlanSchema)))
		})

		It("succeeds with arguments from stdin", func() {
			rawInputParams := serviceadapter.InputParams{
				GeneratePlanSchemas: serviceadapter.GeneratePlanSchemasParams{
					Plan: planJSON,
				},
			}
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
			fakeSchemaGenerator.GeneratePlanSchemaReturns(expectedPlanSchema, nil)

			fakeStdin := bytes.NewBufferString(toJson(rawInputParams))
			err := handler.Handle([]string{commandName, "generate-plan-schemas"}, outputBuffer, errorBuffer, fakeStdin)

			Expect(err).NotTo(HaveOccurred())
			Expect(fakeSchemaGenerator.GeneratePlanSchemaCallCount()).To(Equal(1))

			Expect(fakeSchemaGenerator.GeneratePlanSchemaArgsForCall(0)).To(Equal(plan))

			contents, err := ioutil.ReadAll(outputBuffer)
			Expect(err).NotTo(HaveOccurred())
			Expect(contents).To(MatchJSON(toJson(expectedPlanSchema)))

		})

		It("returns a not-implemented error when there is no generate-plan-schemas handler", func() {
			handler.SchemaGenerator = nil
			err := handler.Handle([]string{commandName, "generate-plan-schemas"}, outputBuffer, errorBuffer, bytes.NewBufferString(""))

			Expect(err).To(BeACLIError(serviceadapter.NotImplementedExitCode, "generate-plan-schemas not implemented"))
		})

		It("returns a missing args error when arguments are missing", func() {
			err := handler.Handle([]string{
				commandName, "generate-plan-schemas", "-plan-json", "",
			}, outputBuffer, errorBuffer, bytes.NewBufferString(""))

			Expect(err).To(BeACLIError(serviceadapter.ErrorExitCode, `Missing arguments for generate-plan-schemas. Usage:`))
		})

		It("returns an error when parsing the arguments fails", func() {
			err := handler.Handle([]string{
				commandName, "generate-plan-schemas", "-plan-json", "not-json",
			}, outputBuffer, errorBuffer, bytes.NewBufferString(""))

			Expect(err).To(MatchError(ContainSubstring("unmarshalling plan JSON")))
		})
	})
})
