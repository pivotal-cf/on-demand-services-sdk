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

			assertCLIHandlerErr(err, serviceadapter.NotImplementedExitCode, "generate-manifest not implemented")
		})

		It("returns a missing args error when arguments are missing", func() {
			err := handler.Handle([]string{
				commandName, "generate-manifest", serviceDeploymentJSON,
			}, outputBuffer, errorBuffer, bytes.NewBufferString(""))

			assertCLIHandlerErr(err, serviceadapter.ErrorExitCode, `Missing arguments for generate-manifest. Usage:`)
		})

		It("returns an error when parsing the arguments fails", func() {
			err := handler.Handle([]string{
				commandName, "generate-manifest", serviceDeploymentJSON, "asd", argsJSON, previousManifestYAML, previousPlanJSON,
			}, outputBuffer, errorBuffer, bytes.NewBufferString(""))

			Expect(err).To(MatchError(ContainSubstring("unmarshalling service plan")))
		})

	})

	Context("When adapter is called with positional arguments", func() {
		It("when no arguments passed returns error", func() {
			err := handler.Handle([]string{commandName}, outputBuffer, errorBuffer, bytes.NewBufferString(""))

			assertCLIHandlerErr(
				err,
				serviceadapter.ErrorExitCode,
				"the following commands are supported: generate-manifest, create-binding, delete-binding, dashboard-url, generate-plan-schemas",
			)
		})

		It("does not output optional commands if not implemented", func() {
			handler.DashboardURLGenerator = nil
			handler.SchemaGenerator = nil
			err := handler.Handle([]string{commandName}, outputBuffer, errorBuffer, bytes.NewBufferString(""))

			assertCLIHandlerErr(
				err,
				serviceadapter.ErrorExitCode,
				"the following commands are supported: generate-manifest, create-binding, delete-binding",
			)
			Expect(err.Error()).NotTo(ContainSubstring("dashboard-url"))
			Expect(err.Error()).NotTo(ContainSubstring("generate-plan-schemas"))
		})

		Describe("Create Binding", func() {
			It("calls the supplied handler passing args through", func() {
				fakeBinder.CreateBindingReturns(expectedBinding, nil)

				err := handler.Handle([]string{
					commandName, "create-binding", bindingID, boshVMsJSON, previousManifestYAML, requestParamsJSON,
				}, outputBuffer, errorBuffer, bytes.NewBufferString(""))

				Expect(err).NotTo(HaveOccurred())

				Expect(fakeBinder.CreateBindingCallCount()).To(Equal(1))
				actualBindingId, actualBoshVMs, actualManifest, actualRequestParams :=
					fakeBinder.CreateBindingArgsForCall(0)

				Expect(actualBindingId).To(Equal(bindingID))
				Expect(actualBoshVMs).To(Equal(boshVMs))
				Expect(actualManifest).To(Equal(previousManifest))
				Expect(actualRequestParams).To(Equal(requestParams))

				Expect(outputBuffer).To(gbytes.Say(toJson(expectedBinding)))
			})

			It("returns a not-implemented error where there is no binder handler", func() {
				handler.Binder = nil
				err := handler.Handle([]string{
					commandName, "create-binding", bindingID, boshVMsJSON, previousManifestYAML, requestParamsJSON,
				}, outputBuffer, errorBuffer, bytes.NewBufferString(""))

				assertCLIHandlerErr(err, serviceadapter.NotImplementedExitCode, "binder not implemented")
			})

			It("returns a missing args error when request JSON is missing", func() {
				err := handler.Handle([]string{
					commandName, "create-binding", bindingID, boshVMsJSON, previousManifestYAML,
				}, outputBuffer, errorBuffer, bytes.NewBufferString(""))

				assertCLIHandlerErr(err, serviceadapter.ErrorExitCode, `Missing arguments for create-binding. Usage:`)
			})

			It("fails with an error when BOSH VMs is corrupt", func() {
				boshVMsJSON += `aaa`
				err := handler.Handle([]string{
					commandName, "create-binding", bindingID, boshVMsJSON, previousManifestYAML, requestParamsJSON,
				}, outputBuffer, errorBuffer, bytes.NewBufferString(""))

				Expect(err).To(MatchError(ContainSubstring("unmarshalling BOSH VMs")))
			})

			It("fails with an error when previous manifest is corrupt", func() {
				previousManifestYAML = previousPlanJSON
				err := handler.Handle([]string{
					commandName, "create-binding", bindingID, boshVMsJSON, previousManifestYAML, requestParamsJSON,
				}, outputBuffer, errorBuffer, bytes.NewBufferString(""))

				Expect(err).To(MatchError(ContainSubstring("unmarshalling manifest YAML")))
			})

			It("fails with an error when request binding params are corrupt", func() {
				requestParamsJSON += "asdf"
				err := handler.Handle([]string{
					commandName, "create-binding", bindingID, boshVMsJSON, previousManifestYAML, requestParamsJSON,
				}, outputBuffer, errorBuffer, bytes.NewBufferString(""))

				Expect(err).To(MatchError(ContainSubstring("unmarshalling request binding parameters")))
			})

			It("returns an error when the binding cannot be created because of generic error", func() {
				fakeBinder.CreateBindingReturns(serviceadapter.Binding{}, errors.New("oops"))
				err := handler.Handle([]string{
					commandName, "create-binding", bindingID, boshVMsJSON, previousManifestYAML, requestParamsJSON,
				}, outputBuffer, errorBuffer, bytes.NewBufferString(""))

				assertCLIHandlerErr(err, serviceadapter.ErrorExitCode, "oops")
				Expect(outputBuffer).To(gbytes.Say("oops"))
			})

			It("returns an error when the binding cannot be created because binding already exists", func() {
				fakeBinder.CreateBindingReturns(serviceadapter.Binding{}, serviceadapter.NewBindingAlreadyExistsError(errors.New("binding already exists")))
				err := handler.Handle([]string{
					commandName, "create-binding", bindingID, boshVMsJSON, previousManifestYAML, requestParamsJSON,
				}, outputBuffer, errorBuffer, bytes.NewBufferString(""))

				assertCLIHandlerErr(err, serviceadapter.BindingAlreadyExistsErrorExitCode, "binding already exists")
				Expect(outputBuffer).To(gbytes.Say("binding already exists"))
			})

			It("returns an error when the binding cannot be created because app guid not provided", func() {
				fakeBinder.CreateBindingReturns(serviceadapter.Binding{}, serviceadapter.NewAppGuidNotProvidedError(errors.New("app guid not provided")))
				err := handler.Handle([]string{
					commandName, "create-binding", bindingID, boshVMsJSON, previousManifestYAML, requestParamsJSON,
				}, outputBuffer, errorBuffer, bytes.NewBufferString(""))

				assertCLIHandlerErr(err, serviceadapter.AppGuidNotProvidedErrorExitCode, "app guid not provided")
				Expect(outputBuffer).To(gbytes.Say("app guid not provided"))
			})
		})

		Describe("Delete Binding", func() {
			It("calls the supplied handler passing args through", func() {
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

			It("returns a not-implemented error where there is no binder handler", func() {
				handler.Binder = nil
				err := handler.Handle([]string{
					commandName, "delete-binding", bindingID, boshVMsJSON, previousManifestYAML, requestParamsJSON,
				}, outputBuffer, errorBuffer, bytes.NewBufferString(""))

				assertCLIHandlerErr(err, serviceadapter.NotImplementedExitCode, "binder not implemented")
			})

			It("returns a missing args error when request JSON is missing", func() {
				err := handler.Handle([]string{
					commandName, "delete-binding", bindingID, boshVMsJSON, previousManifestYAML,
				}, outputBuffer, errorBuffer, bytes.NewBufferString(""))

				assertCLIHandlerErr(err, serviceadapter.ErrorExitCode, `Missing arguments for delete-binding. Usage:`)
			})

			It("fails with an error when BOSH VMs is corrupt", func() {
				boshVMsJSON += `aaa`
				err := handler.Handle([]string{
					commandName, "delete-binding", bindingID, boshVMsJSON, previousManifestYAML, requestParamsJSON,
				}, outputBuffer, errorBuffer, bytes.NewBufferString(""))

				Expect(err).To(MatchError(ContainSubstring("unmarshalling BOSH VMs")))
			})

			It("fails with an error when previous manifest is corrupt", func() {
				previousManifestYAML = previousPlanJSON
				err := handler.Handle([]string{
					commandName, "delete-binding", bindingID, boshVMsJSON, previousManifestYAML, requestParamsJSON,
				}, outputBuffer, errorBuffer, bytes.NewBufferString(""))

				Expect(err).To(MatchError(ContainSubstring("unmarshalling manifest YAML")))
			})

			It("fails with an error when request binding params are corrupt", func() {
				requestParamsJSON += "asdf"
				err := handler.Handle([]string{
					commandName, "delete-binding", bindingID, boshVMsJSON, previousManifestYAML, requestParamsJSON,
				}, outputBuffer, errorBuffer, bytes.NewBufferString(""))

				Expect(err).To(MatchError(ContainSubstring("unmarshalling request binding parameters")))
			})

			It("returns an error when the binding cannot be deleted because of generic error", func() {
				fakeBinder.DeleteBindingReturns(errors.New("oops"))
				err := handler.Handle([]string{
					commandName, "delete-binding", bindingID, boshVMsJSON, previousManifestYAML, requestParamsJSON,
				}, outputBuffer, errorBuffer, bytes.NewBufferString(""))

				assertCLIHandlerErr(err, serviceadapter.ErrorExitCode, "oops")
				Expect(outputBuffer).To(gbytes.Say("oops"))
			})

			It("returns an error when the binding cannot be deleted because binding is not found", func() {
				fakeBinder.DeleteBindingReturns(serviceadapter.NewBindingNotFoundError(errors.New("binding not found")))
				err := handler.Handle([]string{
					commandName, "delete-binding", bindingID, boshVMsJSON, previousManifestYAML, requestParamsJSON,
				}, outputBuffer, errorBuffer, bytes.NewBufferString(""))

				assertCLIHandlerErr(err, serviceadapter.BindingNotFoundErrorExitCode, "binding not found")
				Expect(outputBuffer).To(gbytes.Say("binding not found"))
			})
		})

		Describe("Dashboard URL", func() {
			It("calls the supplied handler passing args through", func() {
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
				Expect(outputBuffer).To(gbytes.Say("http://url.example.com"))
			})

			It("returns a not-implemented error where there is no dashboard-url handler", func() {
				handler.DashboardURLGenerator = nil
				err := handler.Handle([]string{
					commandName, "dashboard-url", instanceID, planJSON, previousManifestYAML,
				}, outputBuffer, errorBuffer, bytes.NewBufferString(""))

				assertCLIHandlerErr(err, serviceadapter.NotImplementedExitCode, "dashboard-url not implemented")
			})

			It("returns a missing args error when the manifest is missing", func() {
				err := handler.Handle([]string{
					commandName, "dashboard-url", instanceID, planJSON,
				}, outputBuffer, errorBuffer, bytes.NewBufferString(""))

				assertCLIHandlerErr(err, serviceadapter.ErrorExitCode, `Missing arguments for dashboard-url. Usage:`)
			})

			It("fails with an error when planJSON is corrupt", func() {
				planJSON += `aaa`
				err := handler.Handle([]string{
					commandName, "dashboard-url", instanceID, planJSON, previousManifestYAML,
				}, outputBuffer, errorBuffer, bytes.NewBufferString(""))

				Expect(err).To(MatchError(ContainSubstring("unmarshalling service plan")))
			})

			It("fails with an error when planJSON is invalid", func() {
				planJSON = `{}`
				err := handler.Handle([]string{
					commandName, "dashboard-url", instanceID, planJSON, previousManifestYAML,
				}, outputBuffer, errorBuffer, bytes.NewBufferString(""))

				Expect(err).To(MatchError(ContainSubstring("validating service plan")))
			})

			It("fails with an error when previous manifest is corrupt", func() {
				previousManifestYAML = previousPlanJSON
				err := handler.Handle([]string{
					commandName, "dashboard-url", instanceID, planJSON, previousManifestYAML,
				}, outputBuffer, errorBuffer, bytes.NewBufferString(""))

				Expect(err).To(MatchError(ContainSubstring("unmarshalling manifest")))
			})

			It("returns an error when the dashboard URL generator fails", func() {
				fakeDashboardUrlGenerator.DashboardUrlReturns(serviceadapter.DashboardUrl{}, errors.New("oops"))
				err := handler.Handle([]string{
					commandName, "dashboard-url", instanceID, planJSON, previousManifestYAML,
				}, outputBuffer, errorBuffer, bytes.NewBufferString(""))

				assertCLIHandlerErr(err, serviceadapter.ErrorExitCode, "oops")
				Expect(outputBuffer).To(gbytes.Say("oops"))
			})
		})

		Describe("Generate Plan Schemas", func() {
			It("returns a not-implemented error where there is no generate-plan-schemas handler", func() {
				handler.SchemaGenerator = nil
				err := handler.Handle([]string{
					commandName, "generate-plan-schemas", "--plan-json", planJSON,
				}, outputBuffer, errorBuffer, bytes.NewBufferString(""))

				assertCLIHandlerErr(err, serviceadapter.NotImplementedExitCode, "plan schema generator not implemented")
			})

			It("returns a plan schema when configured with an schema generator", func() {
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

			It("returns an error if cannot generate the schema for the plan", func() {
				fakeSchemaGenerator.GeneratePlanSchemaReturns(serviceadapter.PlanSchema{}, errors.New("oops"))

				err := handler.Handle([]string{
					commandName, "generate-plan-schemas", "--plan-json", planJSON,
				}, outputBuffer, errorBuffer, bytes.NewBufferString(""))

				assertCLIHandlerErr(err, serviceadapter.ErrorExitCode, "oops")
				Expect(outputBuffer).To(gbytes.Say("oops"))
			})

			It("returns an error if the plan JSON is corrupt", func() {
				planJSON += "asd"
				err := handler.Handle([]string{
					commandName, "generate-plan-schemas", "--plan-json", planJSON,
				}, outputBuffer, errorBuffer, bytes.NewBufferString(""))

				Expect(err).To(MatchError(ContainSubstring("unmarshalling plan JSON")))
			})

			It("returns an error if the plan JSON is invalid", func() {
				err := handler.Handle([]string{
					commandName, "generate-plan-schemas", "--plan-json", `{}`,
				}, outputBuffer, errorBuffer, bytes.NewBufferString(""))

				Expect(err).To(MatchError(ContainSubstring("error validating plan JSON")))
			})

			It("returns an error if the args are not correct", func() {
				err := handler.Handle([]string{
					commandName, "generate-plan-schemas", "a", planJSON,
				}, outputBuffer, errorBuffer, bytes.NewBufferString(""))

				assertCLIHandlerErr(
					err,
					serviceadapter.ErrorExitCode,
					"Incorrect arguments for generate-plan-schemas",
				)
			})

			It("prints a help message, without failing", func() {
				err := handler.Handle([]string{
					commandName, "generate-plan-schemas", "-help",
				}, outputBuffer, errorBuffer, bytes.NewBufferString(""))

				assertCLIHandlerErr(
					err,
					serviceadapter.ErrorExitCode,
					"Incorrect arguments for generate-plan-schemas",
				)
				Expect(errorBuffer).To(gbytes.Say("Usage:"))
			})

			It("prints a help message, without failing, even when plan JSON is set", func() {
				err := handler.Handle([]string{
					commandName, "generate-plan-schemas", "--plan-json", "{}", "-help",
				}, outputBuffer, errorBuffer, bytes.NewBufferString(""))

				assertCLIHandlerErr(
					err,
					serviceadapter.ErrorExitCode,
					"Incorrect arguments for generate-plan-schemas",
				)
				Expect(errorBuffer).To(gbytes.Say("Usage:"))
			})
		})
	})

	Context("When adapter passes arguments through STDIN", func() {
		var (
			subCommand     string
			command        []string
			fakeStdin      io.Reader
			rawInputParams serviceadapter.InputParams
		)

		BeforeEach(func() {
			var err error
			Expect(err).NotTo(HaveOccurred())

			outputBuffer = gbytes.NewBuffer()
			errorBuffer = gbytes.NewBuffer()

			handler = serviceadapter.CommandLineHandler{
				ManifestGenerator:     fakeManifestGenerator,
				Binder:                fakeBinder,
				DashboardURLGenerator: fakeDashboardUrlGenerator,
				SchemaGenerator:       fakeSchemaGenerator,
			}
		})

		JustBeforeEach(func() {
			command = []string{commandName, subCommand, "-stdin"}

			fakeStdin = bytes.NewBuffer([]byte(toJson(rawInputParams)))
		})

		Describe("dashboard-url", func() {
			BeforeEach(func() {
				subCommand = "dashboard-url"

				rawInputParams = serviceadapter.InputParams{
					DashboardUrl: serviceadapter.DashboardUrlParams{
						InstanceId: instanceID,
						Plan:       toJson(plan),
						Manifest:   previousManifestYAML,
					},
				}
			})

			It("calls the supplied handler passing args through", func() {
				fakeDashboardUrlGenerator.DashboardUrlReturns(serviceadapter.DashboardUrl{DashboardUrl: "http://url.example.com"}, nil)

				err := handler.Handle(command, outputBuffer, errorBuffer, fakeStdin)
				Expect(err).NotTo(HaveOccurred())

				Expect(fakeDashboardUrlGenerator.DashboardUrlCallCount()).To(Equal(1))
				actualInstanceID, actualPlanJSON, actualManifestYAML := fakeDashboardUrlGenerator.DashboardUrlArgsForCall(0)

				Expect(actualInstanceID).To(Equal(instanceID))
				Expect(actualPlanJSON).To(Equal(plan))
				Expect(actualManifestYAML).To(Equal(previousManifest))
				Expect(outputBuffer).To(gbytes.Say("http://url.example.com"))
			})

			It("returns a not-implemented error where there is no dashboard-url handler", func() {
				handler.DashboardURLGenerator = nil
				err := handler.Handle(command, outputBuffer, errorBuffer, fakeStdin)
				assertCLIHandlerErr(err, serviceadapter.NotImplementedExitCode, "dashboard-url not implemented")
			})

			Describe("error handling", func() {
				It("fails with an error when planJSON is corrupt", func() {
					rawInputParams.DashboardUrl.Plan = ""
					fakeStdin = bytes.NewBufferString(toJson(rawInputParams))
					err := handler.Handle(command, outputBuffer, errorBuffer, fakeStdin)
					Expect(err).To(MatchError(ContainSubstring("unmarshalling service plan")))
				})

				It("fails with an error when planJSON is invalid", func() {
					rawInputParams.DashboardUrl.Plan = "{}"
					fakeStdin = bytes.NewBufferString(toJson(rawInputParams))
					err := handler.Handle(command, outputBuffer, errorBuffer, fakeStdin)
					Expect(err).To(MatchError(ContainSubstring("validating service plan")))
				})

				It("returns an error if cannot read from stdin", func() {
					fakeStdin = NewFakeReader()
					err := handler.Handle(command, outputBuffer, errorBuffer, fakeStdin)

					assertCLIHandlerErr(err, serviceadapter.ErrorExitCode, "error reading input params JSON")
				})

				It("fails with an error when previous manifest is corrupt", func() {
					rawInputParams.DashboardUrl.Manifest = "not a manifest"
					fakeStdin = bytes.NewBufferString(toJson(rawInputParams))
					err := handler.Handle(command, outputBuffer, errorBuffer, fakeStdin)
					Expect(err).To(MatchError(ContainSubstring("unmarshalling manifest")))
				})

				It("returns an error when the dashboard URL generator fails", func() {
					fakeDashboardUrlGenerator.DashboardUrlReturns(serviceadapter.DashboardUrl{}, errors.New("oops"))
					err := handler.Handle(command, outputBuffer, errorBuffer, fakeStdin)

					assertCLIHandlerErr(err, serviceadapter.ErrorExitCode, "oops")
					Expect(outputBuffer).To(gbytes.Say("oops"))
				})
			})
		})

		Describe("create-binding", func() {
			BeforeEach(func() {
				subCommand = "create-binding"

				rawInputParams = serviceadapter.InputParams{
					CreateBinding: serviceadapter.CreateBindingParams{
						RequestParameters: toJson(requestParams),
						BindingId:         bindingID,
						BoshVms:           toJson(boshVMs),
						Manifest:          toYaml(previousManifest),
					},
				}
			})

			It("calls the supplied handler passing args through", func() {
				fakeBinder.CreateBindingReturns(expectedBinding, nil)

				err := handler.Handle(command, outputBuffer, errorBuffer, fakeStdin)
				Expect(err).NotTo(HaveOccurred())

				Expect(fakeBinder.CreateBindingCallCount()).To(Equal(1))

				actualBindingId, actualBoshVMs, actualManifest, actualRequestParams :=
					fakeBinder.CreateBindingArgsForCall(0)

				Expect(actualBindingId).To(Equal(bindingID))
				Expect(actualBoshVMs).To(Equal(boshVMs))
				Expect(actualManifest).To(Equal(previousManifest))
				Expect(actualRequestParams).To(Equal(requestParams))

				Expect(outputBuffer).To(gbytes.Say(toJson(expectedBinding)))
			})

			It("returns a not-implemented error where there is no binder handler", func() {
				handler.Binder = nil
				err := handler.Handle(command, outputBuffer, errorBuffer, fakeStdin)
				assertCLIHandlerErr(err, serviceadapter.NotImplementedExitCode, "binder not implemented")
			})

			Describe("error handling", func() {
				It("fails with an error when BOSH VMs is corrupt", func() {
					rawInputParams.CreateBinding.BoshVms = "foo"
					fakeStdin = bytes.NewBufferString(toJson(rawInputParams))
					err := handler.Handle(command, outputBuffer, errorBuffer, fakeStdin)

					Expect(err).To(MatchError(ContainSubstring("unmarshalling BOSH VMs")))
				})

				It("fails with an error when previous manifest is corrupt", func() {
					rawInputParams.CreateBinding.Manifest = "foo"
					fakeStdin = bytes.NewBufferString(toJson(rawInputParams))
					err := handler.Handle(command, outputBuffer, errorBuffer, fakeStdin)

					Expect(err).To(MatchError(ContainSubstring("unmarshalling manifest YAML")))
				})

				It("fails with an error when request binding params are corrupt", func() {
					rawInputParams.CreateBinding.RequestParameters = "asdf"
					fakeStdin = bytes.NewBufferString(toJson(rawInputParams))
					err := handler.Handle(command, outputBuffer, errorBuffer, fakeStdin)

					Expect(err).To(MatchError(ContainSubstring("unmarshalling request binding parameters")))
				})

				It("returns an error if cannot read from stdin", func() {
					fakeStdin = NewFakeReader()
					err := handler.Handle(command, outputBuffer, errorBuffer, fakeStdin)

					assertCLIHandlerErr(err, serviceadapter.ErrorExitCode, "error reading input params JSON")
				})

				It("returns an error when the binding cannot be created because of generic error", func() {
					fakeBinder.CreateBindingReturns(serviceadapter.Binding{}, errors.New("oops"))
					fakeStdin = bytes.NewBufferString(toJson(rawInputParams))
					err := handler.Handle(command, outputBuffer, errorBuffer, fakeStdin)

					assertCLIHandlerErr(err, serviceadapter.ErrorExitCode, "oops")
					Expect(outputBuffer).To(gbytes.Say("oops"))
				})

				It("returns an error when the binding cannot be created because binding already exists", func() {
					fakeBinder.CreateBindingReturns(serviceadapter.Binding{}, serviceadapter.NewBindingAlreadyExistsError(errors.New("binding already exists")))
					fakeStdin = bytes.NewBufferString(toJson(rawInputParams))
					err := handler.Handle(command, outputBuffer, errorBuffer, fakeStdin)

					assertCLIHandlerErr(err, serviceadapter.BindingAlreadyExistsErrorExitCode, "binding already exists")
					Expect(outputBuffer).To(gbytes.Say("binding already exists"))
				})

				It("returns an error when the binding cannot be created because app guid not provided", func() {
					fakeBinder.CreateBindingReturns(serviceadapter.Binding{}, serviceadapter.NewAppGuidNotProvidedError(errors.New("app guid not provided")))
					fakeStdin = bytes.NewBufferString(toJson(rawInputParams))
					err := handler.Handle(command, outputBuffer, errorBuffer, fakeStdin)

					assertCLIHandlerErr(err, serviceadapter.AppGuidNotProvidedErrorExitCode, "app guid not provided")
					Expect(outputBuffer).To(gbytes.Say("app guid not provided"))
				})
			})
		})

		Describe("delete Binding", func() {
			BeforeEach(func() {
				subCommand = "delete-binding"

				rawInputParams = serviceadapter.InputParams{
					DeleteBinding: serviceadapter.DeleteBindingParams{
						RequestParameters: toJson(requestParams),
						BindingId:         bindingID,
						BoshVms:           toJson(boshVMs),
						Manifest:          toYaml(previousManifest),
					},
				}
			})

			It("calls the supplied handler passing args through", func() {
				fakeBinder.DeleteBindingReturns(nil)

				err := handler.Handle(command, outputBuffer, errorBuffer, fakeStdin)
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
				err := handler.Handle(command, outputBuffer, errorBuffer, fakeStdin)
				assertCLIHandlerErr(err, serviceadapter.NotImplementedExitCode, "binder not implemented")
			})

			Describe("error handling", func() {
				It("fails with an error when BOSH VMs is corrupt", func() {
					rawInputParams.DeleteBinding.BoshVms = "notjson"
					fakeStdin = bytes.NewBufferString(toJson(rawInputParams))
					err := handler.Handle(command, outputBuffer, errorBuffer, fakeStdin)
					Expect(err).To(MatchError(ContainSubstring("unmarshalling BOSH VMs")))
				})

				It("fails with an error when previous manifest is corrupt", func() {
					rawInputParams.DeleteBinding.Manifest = "notyaml"
					fakeStdin = bytes.NewBufferString(toJson(rawInputParams))
					err := handler.Handle(command, outputBuffer, errorBuffer, fakeStdin)
					Expect(err).To(MatchError(ContainSubstring("unmarshalling manifest YAML")))
				})

				It("fails with an error when request binding params are corrupt", func() {
					rawInputParams.DeleteBinding.RequestParameters = "notjson"
					fakeStdin = bytes.NewBufferString(toJson(rawInputParams))
					err := handler.Handle(command, outputBuffer, errorBuffer, fakeStdin)
					Expect(err).To(MatchError(ContainSubstring("unmarshalling request binding parameters")))
				})

				It("returns an error if cannot read from stdin", func() {
					fakeStdin = NewFakeReader()
					err := handler.Handle(command, outputBuffer, errorBuffer, fakeStdin)

					assertCLIHandlerErr(err, serviceadapter.ErrorExitCode, "error reading input params JSON")
				})

				It("returns an error when the binding cannot be deleted because of generic error", func() {
					fakeBinder.DeleteBindingReturns(errors.New("oops"))
					err := handler.Handle(command, outputBuffer, errorBuffer, fakeStdin)

					assertCLIHandlerErr(err, serviceadapter.ErrorExitCode, "oops")
					Expect(outputBuffer).To(gbytes.Say("oops"))
				})

				It("returns an error when the binding cannot be deleted because binding is not found", func() {
					fakeBinder.DeleteBindingReturns(serviceadapter.NewBindingNotFoundError(errors.New("binding not found")))
					err := handler.Handle(command, outputBuffer, errorBuffer, fakeStdin)

					assertCLIHandlerErr(err, serviceadapter.BindingNotFoundErrorExitCode, "binding not found")
					Expect(outputBuffer).To(gbytes.Say("binding not found"))
				})
			})
		})

		Describe("generate-plan-schema", func() {
			BeforeEach(func() {
				subCommand = "generate-plan-schemas"

				rawInputParams = serviceadapter.InputParams{
					GeneratePlanSchemas: serviceadapter.GeneratePlanSchemasParams{
						Plan: planJSON,
					},
				}
			})

			It("returns a not-implemented error where there is no generate-plan-schemas handler", func() {
				handler.SchemaGenerator = nil
				err := handler.Handle(command, outputBuffer, errorBuffer, bytes.NewBufferString(""))

				assertCLIHandlerErr(err, serviceadapter.NotImplementedExitCode, "plan schema generator not implemented")
			})

			It("returns a plan schema when configured with an schema generator", func() {
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

				err := handler.Handle(command, outputBuffer, errorBuffer, fakeStdin)

				Expect(err).NotTo(HaveOccurred())
				Expect(fakeSchemaGenerator.GeneratePlanSchemaCallCount()).To(Equal(1))

				Expect(fakeSchemaGenerator.GeneratePlanSchemaArgsForCall(0)).To(Equal(plan))

				contents, err := ioutil.ReadAll(outputBuffer)
				Expect(err).NotTo(HaveOccurred())
				Expect(contents).To(MatchJSON(toJson(expectedPlanSchema)))
			})

			It("returns an error if cannot generate the schema for the plan", func() {
				fakeSchemaGenerator.GeneratePlanSchemaReturns(serviceadapter.PlanSchema{}, errors.New("oops"))

				err := handler.Handle(command, outputBuffer, errorBuffer, fakeStdin)

				assertCLIHandlerErr(err, serviceadapter.ErrorExitCode, "oops")
				Expect(outputBuffer).To(gbytes.Say("oops"))
			})

			It("returns an error if cannot read from stdin", func() {
				fakeStdin = NewFakeReader()
				err := handler.Handle(command, outputBuffer, errorBuffer, fakeStdin)

				assertCLIHandlerErr(err, serviceadapter.ErrorExitCode, "error reading input params JSON")
			})

			It("returns an error if the plan JSON is corrupt", func() {
				rawInputParams.GeneratePlanSchemas.Plan = "notjson"
				fakeStdin = bytes.NewBufferString(toJson(rawInputParams))
				err := handler.Handle(command, outputBuffer, errorBuffer, fakeStdin)

				Expect(err).To(MatchError(ContainSubstring("unmarshalling plan JSON")))
			})

			It("returns an error if the plan JSON is invalid", func() {
				rawInputParams.GeneratePlanSchemas.Plan = "{}"
				fakeStdin = bytes.NewBufferString(toJson(rawInputParams))
				err := handler.Handle(command, outputBuffer, errorBuffer, fakeStdin)

				Expect(err).To(MatchError(ContainSubstring("error validating plan JSON")))
			})
		})
	})
})

func assertCLIHandlerErr(err error, exitCode int, message string) {
	Expect(err).To(HaveOccurred())
	Expect(err).To(BeAssignableToTypeOf(serviceadapter.CLIHandlerError{}))
	Expect(err).To(MatchError(ContainSubstring(message)))
	actualErr := err.(serviceadapter.CLIHandlerError)
	Expect(actualErr.ExitCode).To(Equal(exitCode))
}

func NewFakeReader() io.Reader {
	return &FakeReader{}
}

type FakeReader struct {
}

func (f *FakeReader) Read(b []byte) (int, error) {
	return 1, fmt.Errorf("fool!")
}
