package serviceadapter_test

import (
	"bytes"
	"encoding/json"
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/pivotal-cf/on-demand-services-sdk/serviceadapter"
	"github.com/pivotal-cf/on-demand-services-sdk/serviceadapter/fakes"
)

var _ = Describe("GeneratePlanSchemas", func() {
	var (
		fakeSchemaGenerator *fakes.FakeSchemaGenerator
		plan                serviceadapter.Plan
		expectedInputParams serviceadapter.InputParams
		action              *serviceadapter.GeneratePlanSchemasAction
		outputBuffer        *gbytes.Buffer
		errorBuffer         *gbytes.Buffer
	)

	BeforeEach(func() {
		fakeSchemaGenerator = new(fakes.FakeSchemaGenerator)
		plan = defaultPlan()
		outputBuffer = gbytes.NewBuffer()
		errorBuffer = gbytes.NewBuffer()

		expectedInputParams = serviceadapter.InputParams{
			GeneratePlanSchemas: serviceadapter.GeneratePlanSchemasParams{
				Plan: toJson(plan),
			},
		}

		action = serviceadapter.NewGeneratePlanSchemasAction(fakeSchemaGenerator, errorBuffer)
	})

	Describe("IsImplemented", func() {
		It("returns true if implemented", func() {
			g := serviceadapter.NewGeneratePlanSchemasAction(fakeSchemaGenerator, errorBuffer)
			Expect(g.IsImplemented()).To(BeTrue())
		})

		It("returns false if not implemented", func() {
			g := serviceadapter.NewGeneratePlanSchemasAction(nil, errorBuffer)
			Expect(g.IsImplemented()).To(BeFalse())
		})
	})

	Describe("ParseArgs", func() {
		When("giving arguments in stdin", func() {
			It("can parse arguments from stdin", func() {
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
			It("can parse positional arguments", func() {
				positionalArgs := []string{
					"-plan-json",
					expectedInputParams.GeneratePlanSchemas.Plan,
				}

				actualInputParams, err := action.ParseArgs(nil, positionalArgs)
				Expect(err).NotTo(HaveOccurred())
				expectedInputParams.TextOutput = true
				Expect(actualInputParams).To(Equal(expectedInputParams))
			})

			It("returns an error when required arguments are not passed in", func() {
				_, err := action.ParseArgs(nil, []string{"-plan-json", ""})
				Expect(err).To(HaveOccurred())
				Expect(err).To(BeAssignableToTypeOf(serviceadapter.MissingArgsError{}))
				Expect(err).To(MatchError(ContainSubstring("<plan-JSON>")))
			})

			It("returns an error when unrecognised arguments are passed in", func() {
				_, err := action.ParseArgs(nil, []string{"-what"})
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError(ContainSubstring("flag provided but not defined: -what")))
			})
		})
	})

	Describe("Execute", func() {
		It("calls the supplied handler passing args through", func() {
			planSchema := serviceadapter.PlanSchema{
				ServiceInstance: serviceadapter.ServiceInstanceSchema{
					Create: serviceadapter.JSONSchemas{
						Parameters: map[string]interface{}{
							"foo": "string",
						},
					},
				},
			}
			fakeSchemaGenerator.GeneratePlanSchemaReturns(planSchema, nil)

			err := action.Execute(expectedInputParams, outputBuffer)
			Expect(err).NotTo(HaveOccurred())

			Expect(fakeSchemaGenerator.GeneratePlanSchemaCallCount()).To(Equal(1))
			actualPlan := fakeSchemaGenerator.GeneratePlanSchemaArgsForCall(0)

			Expect(actualPlan).To(Equal(plan))
			var planSchemasOutput serviceadapter.PlanSchema
			Expect(json.Unmarshal(outputBuffer.Contents(), &planSchemasOutput)).To(Succeed())
			Expect(planSchemasOutput).To(Equal(planSchema))
		})

		Context("error handling", func() {
			It("returns an error when plan cannot be unmarshalled", func() {
				expectedInputParams.GeneratePlanSchemas.Plan = "not-json"
				err := action.Execute(expectedInputParams, outputBuffer)
				Expect(err).To(MatchError(ContainSubstring("unmarshalling plan JSON")))
			})

			It("returns an error when plan is invalid", func() {
				expectedInputParams.GeneratePlanSchemas.Plan = "{}"
				err := action.Execute(expectedInputParams, outputBuffer)
				Expect(err).To(MatchError(ContainSubstring("validating plan JSON")))
			})

			It("returns an error when schemaGenerator returns an error", func() {
				fakeSchemaGenerator.GeneratePlanSchemaReturns(serviceadapter.PlanSchema{}, errors.New("something went wrong"))
				err := action.Execute(expectedInputParams, outputBuffer)
				Expect(err).To(BeACLIError(1, "something went wrong"))
			})

			It("returns an error when the returned object cannot be unmarshalled", func() {
				fakeWriter := new(FakeWriter)
				fakeSchemaGenerator.GeneratePlanSchemaReturns(serviceadapter.PlanSchema{}, nil)

				err := action.Execute(expectedInputParams, fakeWriter)
				Expect(err).To(MatchError(ContainSubstring("marshalling plan schema")))
			})

		})
	})
})
