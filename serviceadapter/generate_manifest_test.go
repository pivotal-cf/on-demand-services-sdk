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
	"bytes"
	"encoding/json"
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"

	"github.com/pivotal-cf/on-demand-services-sdk/bosh"
	"github.com/pivotal-cf/on-demand-services-sdk/serviceadapter"
	"github.com/pivotal-cf/on-demand-services-sdk/serviceadapter/fakes"
)

var _ = Describe("GenerateManifest", func() {
	var (
		fakeManifestGenerator *fakes.FakeManifestGenerator
		serviceDeployment     serviceadapter.ServiceDeployment
		requestParams         serviceadapter.RequestParameters
		plan                  serviceadapter.Plan
		previousPlan          serviceadapter.Plan
		previousManifest      bosh.BoshManifest
		secretsMap            serviceadapter.ManifestSecrets
		expectedInputParams   serviceadapter.InputParams
		action                *serviceadapter.GenerateManifestAction
		outputBuffer          *gbytes.Buffer
	)

	BeforeEach(func() {
		fakeManifestGenerator = new(fakes.FakeManifestGenerator)
		serviceDeployment = defaultServiceDeployment()
		requestParams = defaultRequestParams()
		plan = defaultPlan()
		previousPlan = defaultPreviousPlan()
		previousManifest = defaultManifest()
		secretsMap = serviceadapter.ManifestSecrets{
			"foo": "b4r",
		}
		outputBuffer = gbytes.NewBuffer()

		expectedInputParams = serviceadapter.InputParams{
			GenerateManifest: serviceadapter.GenerateManifestParams{
				ServiceDeployment: toJson(serviceDeployment),
				Plan:              toJson(plan),
				PreviousPlan:      toJson(previousPlan),
				RequestParameters: toJson(requestParams),
				PreviousManifest:  toYaml(previousManifest),
			},
		}

		action = serviceadapter.NewGenerateManifestAction(fakeManifestGenerator)
	})

	Describe("IsImplemented", func() {
		It("returns true if implemented", func() {
			g := serviceadapter.NewGenerateManifestAction(fakeManifestGenerator)
			Expect(g.IsImplemented()).To(BeTrue())
		})

		It("returns false if not implemented", func() {
			g := serviceadapter.NewGenerateManifestAction(nil)
			Expect(g.IsImplemented()).To(BeFalse())
		})
	})

	Describe("ParseArgs", func() {
		When("giving arguments in stdin", func() {
			It("can parse arguments from stdin", func() {
				expectedInputParams.GenerateManifest.PreviousSecrets = toJson(secretsMap)
				input := bytes.NewBuffer([]byte(toJson(expectedInputParams)))
				actualInputParams, err := action.ParseArgs(input, []string{})

				Expect(err).NotTo(HaveOccurred())
				Expect(actualInputParams).To(Equal(expectedInputParams))
			})

			It("returns an error when cannot read from input buffer", func() {
				fakeReader := new(FakeReader)
				_, err := action.ParseArgs(fakeReader, []string{})
				Expect(err).To(BeACLIError(1, "error reading input params JSON"))
			})

			It("returns an error when cannot unmarshal from input buffer", func() {
				input := bytes.NewBuffer([]byte("not-valid-json"))
				_, err := action.ParseArgs(input, []string{})
				Expect(err).To(BeACLIError(1, "error unmarshalling input params JSON"))
			})

			It("returns an error when input buffer is empty", func() {
				input := bytes.NewBuffer([]byte{})
				_, err := action.ParseArgs(input, []string{})
				Expect(err).To(BeACLIError(1, "expecting parameters to be passed via stdin"))
			})
		})

		When("given positional arguments", func() {
			var emptyBuffer *bytes.Buffer

			BeforeEach(func() {
				emptyBuffer = bytes.NewBuffer(nil)
			})

			It("can parse positional arguments", func() {
				positionalArgs := []string{
					expectedInputParams.GenerateManifest.ServiceDeployment,
					expectedInputParams.GenerateManifest.Plan,
					expectedInputParams.GenerateManifest.RequestParameters,
					expectedInputParams.GenerateManifest.PreviousManifest,
					expectedInputParams.GenerateManifest.PreviousPlan,
				}

				actualInputParams, err := action.ParseArgs(emptyBuffer, positionalArgs)
				Expect(err).NotTo(HaveOccurred())
				expectedInputParams.TextOutput = true
				Expect(actualInputParams).To(Equal(expectedInputParams))
			})

			It("returns an error when required arguments are not passed in", func() {
				_, err := action.ParseArgs(emptyBuffer, []string{"foo"})
				Expect(err).To(HaveOccurred())
				Expect(err).To(BeAssignableToTypeOf(serviceadapter.MissingArgsError{}))
				Expect(err).To(MatchError(ContainSubstring("<service-deployment-JSON> <plan-JSON> <request-params-JSON> <previous-manifest-YAML> <previous-plan-JSON>")))
			})
		})
	})

	Describe("Execute", func() {
		It("calls the supplied handler passing args through", func() {
			manifest := bosh.BoshManifest{Name: "bill"}
			fakeManifestGenerator.GenerateManifestReturns(serviceadapter.GenerateManifestOutput{Manifest: manifest, ODBManagedSecrets: serviceadapter.ODBManagedSecrets{}}, nil)

			expectedInputParams.GenerateManifest.PreviousSecrets = toJson(secretsMap)
			err := action.Execute(expectedInputParams, outputBuffer)

			Expect(err).NotTo(HaveOccurred())

			Expect(fakeManifestGenerator.GenerateManifestCallCount()).To(Equal(1))
			actualServiceDeployment, actualPlan, actualRequestParams, actualPreviousManifest, actualPreviousPlan, actualSecretsMap :=
				fakeManifestGenerator.GenerateManifestArgsForCall(0)

			Expect(actualServiceDeployment).To(Equal(serviceDeployment))
			Expect(actualPlan).To(Equal(plan))
			Expect(actualRequestParams).To(Equal(requestParams))
			Expect(actualPreviousManifest).To(Equal(&previousManifest))
			Expect(actualPreviousPlan).To(Equal(&previousPlan))
			Expect(actualSecretsMap).To(Equal(secretsMap))

			var output serviceadapter.MarshalledGenerateManifest
			Expect(json.Unmarshal(outputBuffer.Contents(), &output)).To(Succeed())
			Expect(output.Manifest).To(Equal(toYaml(manifest)))
		})

		When("not outputting json", func() {
			It("outputs the manifest as text", func() {
				manifest := bosh.BoshManifest{Name: "bill"}
				fakeManifestGenerator.GenerateManifestReturns(serviceadapter.GenerateManifestOutput{Manifest: manifest}, nil)

				expectedInputParams.TextOutput = true
				err := action.Execute(expectedInputParams, outputBuffer)

				Expect(err).NotTo(HaveOccurred())
				Expect(string(outputBuffer.Contents())).To(Equal(toYaml(manifest)))
			})
		})

		Context("error handling", func() {
			It("returns an error when service deployment cannot be unmarshalled", func() {
				expectedInputParams.GenerateManifest.ServiceDeployment = "not-json"
				err := action.Execute(expectedInputParams, outputBuffer)
				Expect(err).To(MatchError(ContainSubstring("unmarshalling service deployment")))
			})

			It("returns an error when service deployment is invalid", func() {
				expectedInputParams.GenerateManifest.ServiceDeployment = "{}"
				err := action.Execute(expectedInputParams, outputBuffer)
				Expect(err).To(MatchError(ContainSubstring("validating service deployment")))
			})

			It("returns an error when plan cannot be unmarshalled", func() {
				expectedInputParams.GenerateManifest.Plan = "not-json"
				err := action.Execute(expectedInputParams, outputBuffer)
				Expect(err).To(MatchError(ContainSubstring("unmarshalling service plan")))
			})

			It("returns an error when plan is invalid", func() {
				expectedInputParams.GenerateManifest.Plan = "{}"
				err := action.Execute(expectedInputParams, outputBuffer)
				Expect(err).To(MatchError(ContainSubstring("validating service plan")))
			})

			It("returns an error when request params cannot be unmarshalled", func() {
				expectedInputParams.GenerateManifest.RequestParameters = "not-json"
				err := action.Execute(expectedInputParams, outputBuffer)
				Expect(err).To(MatchError(ContainSubstring("unmarshalling requestParams")))
			})

			It("returns an error when previous manifest cannot be unmarshalled", func() {
				expectedInputParams.GenerateManifest.PreviousManifest = "not-yaml"
				err := action.Execute(expectedInputParams, outputBuffer)
				Expect(err).To(MatchError(ContainSubstring("unmarshalling previous manifest")))
			})

			It("returns an error when previous plan cannot be unmarshalled", func() {
				expectedInputParams.GenerateManifest.PreviousPlan = "not-json"
				err := action.Execute(expectedInputParams, outputBuffer)
				Expect(err).To(MatchError(ContainSubstring("unmarshalling previous service plan")))
			})

			It("returns an error when previous plan is invalid", func() {
				expectedInputParams.GenerateManifest.PreviousPlan = "{}"
				err := action.Execute(expectedInputParams, outputBuffer)
				Expect(err).To(MatchError(ContainSubstring("validating previous service plan")))
			})

			It("returns an error when manifestGenerator returns an error", func() {
				fakeManifestGenerator.GenerateManifestReturns(serviceadapter.GenerateManifestOutput{}, errors.New("something went wrong"))
				err := action.Execute(expectedInputParams, outputBuffer)
				Expect(err).To(BeACLIError(1, "something went wrong"))
			})

			It("returns an error when the manifest is invalid", func() {
				var invalidYAML struct {
					A int
					B int `yaml:"a"`
				}

				manifest := bosh.BoshManifest{
					Tags: map[string]interface{}{"foo": invalidYAML},
				}

				expectedInputParams.TextOutput = true
				fakeManifestGenerator.GenerateManifestReturns(serviceadapter.GenerateManifestOutput{Manifest: manifest}, nil)
				err := action.Execute(expectedInputParams, outputBuffer)
				Expect(err).To(MatchError(ContainSubstring("error marshalling bosh manifest")))
			})

			It("returns an error when the generated output cannot be marshalled", func() {
				manifest := bosh.BoshManifest{
					Tags: map[string]interface{}{"foo": make(chan int)},
				}

				fakeManifestGenerator.GenerateManifestReturns(serviceadapter.GenerateManifestOutput{Manifest: manifest}, nil)
				err := action.Execute(expectedInputParams, outputBuffer)
				Expect(err).To(MatchError(ContainSubstring("error marshalling bosh manifest")))
			})
		})
	})
})
