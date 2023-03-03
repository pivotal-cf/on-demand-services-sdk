package serviceadapter_test

import (
	"bytes"
	"encoding/json"
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/pivotal-cf/on-demand-services-sdk/bosh"
	"github.com/pivotal-cf/on-demand-services-sdk/serviceadapter"
	"github.com/pivotal-cf/on-demand-services-sdk/serviceadapter/fakes"
)

var _ = Describe("CreateBinding", func() {
	var (
		fakeBinder    *fakes.FakeBinder
		bindingId     string
		boshVMs       bosh.BoshVMs
		requestParams serviceadapter.RequestParameters
		manifest      bosh.BoshManifest
		dnsAddresses  serviceadapter.DNSAddresses

		expectedInputParams serviceadapter.InputParams
		action              *serviceadapter.CreateBindingAction
		outputBuffer        *gbytes.Buffer
	)

	BeforeEach(func() {
		fakeBinder = new(fakes.FakeBinder)
		bindingId = "binding-id"
		boshVMs = bosh.BoshVMs{}
		requestParams = defaultRequestParams()
		manifest = defaultManifest()
		outputBuffer = gbytes.NewBuffer()

		expectedInputParams = serviceadapter.InputParams{
			CreateBinding: serviceadapter.CreateBindingJSONParams{
				BindingId:         bindingId,
				BoshVms:           toJson(boshVMs),
				Manifest:          toYaml(manifest),
				RequestParameters: toJson(requestParams),
			},
		}

		action = serviceadapter.NewCreateBindingAction(fakeBinder)
	})

	Describe("IsImplemented", func() {
		It("returns true if implemented", func() {
			Expect(action.IsImplemented()).To(BeTrue())
		})

		It("returns false if not implemented", func() {
			c := serviceadapter.NewCreateBindingAction(nil)
			Expect(c.IsImplemented()).To(BeFalse())
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

			It("can parse manifest secrets", func() {
				expectedInputParamsWithSecrets := expectedInputParams
				expectedInputParamsWithSecrets.CreateBinding.Secrets = `{ "/foo": "{ "status": "foo" }" }`
				input := bytes.NewBuffer([]byte(toJson(expectedInputParamsWithSecrets)))
				actualInputParams, err := action.ParseArgs(input, []string{})

				Expect(err).NotTo(HaveOccurred())
				Expect(actualInputParams).To(Equal(expectedInputParamsWithSecrets))
			})
		})

		When("given positional arguments", func() {
			It("can parse positional arguments", func() {
				positionalArgs := []string{
					expectedInputParams.CreateBinding.BindingId,
					expectedInputParams.CreateBinding.BoshVms,
					expectedInputParams.CreateBinding.Manifest,
					expectedInputParams.CreateBinding.RequestParameters,
				}

				actualInputParams, err := action.ParseArgs(nil, positionalArgs)
				Expect(err).NotTo(HaveOccurred())
				Expect(actualInputParams).To(Equal(expectedInputParams))
			})

			It("returns an error when required arguments are not passed in", func() {
				_, err := action.ParseArgs(nil, []string{"foo"})
				Expect(err).To(HaveOccurred())
				Expect(err).To(BeAssignableToTypeOf(serviceadapter.MissingArgsError{}))
				Expect(err).To(MatchError(ContainSubstring("<binding-ID> <bosh-VMs-JSON> <manifest-YAML> <request-params-JSON>")))
			})
		})
	})

	Describe("Execute", func() {
		It("calls the supplied handler passing args through", func() {
			binding := serviceadapter.Binding{
				Credentials: map[string]interface{}{
					"password": "letmein",
				},
			}
			fakeBinder.CreateBindingReturns(binding, nil)

			dnsAddresses = serviceadapter.DNSAddresses{
				"config-1": "some-dns.bosh",
			}
			expectedInputParams.CreateBinding.DNSAddresses = toJson(dnsAddresses)
			expectedInputParams.CreateBinding.Secrets = `{ "/foo": "{ \"status\": \"bar\" }" }`
			err := action.Execute(expectedInputParams, outputBuffer)

			Expect(err).NotTo(HaveOccurred())

			Expect(fakeBinder.CreateBindingCallCount()).To(Equal(1))
			params := fakeBinder.CreateBindingArgsForCall(0)

			Expect(params.BindingID).To(Equal(bindingId))
			Expect(params.DeploymentTopology).To(Equal(boshVMs))
			Expect(params.Manifest).To(Equal(manifest))
			Expect(params.RequestParams).To(Equal(requestParams))
			Expect(params.DNSAddresses).To(Equal(dnsAddresses))
			Expect(params.Secrets).To(Equal(serviceadapter.ManifestSecrets{
				"/foo": `{ "status": "bar" }`,
			}))

			var bindingOutput serviceadapter.Binding
			json.Unmarshal(outputBuffer.Contents(), &bindingOutput)
			Expect(bindingOutput).To(Equal(binding))
		})

		Context("error handling", func() {
			It("returns an error when bosh VMs cannot be unmarshalled", func() {
				expectedInputParams.CreateBinding.BoshVms = "not-json"
				err := action.Execute(expectedInputParams, outputBuffer)
				Expect(err).To(MatchError(ContainSubstring("unmarshalling BOSH VMs")))
			})

			It("returns an error when manifest cannot be unmarshalled", func() {
				expectedInputParams.CreateBinding.Manifest = "not-yaml"
				err := action.Execute(expectedInputParams, outputBuffer)
				Expect(err).To(MatchError(ContainSubstring("unmarshalling manifest YAML")))
			})

			It("returns an error when request params cannot be unmarshalled", func() {
				expectedInputParams.CreateBinding.RequestParameters = "not-json"
				err := action.Execute(expectedInputParams, outputBuffer)
				Expect(err).To(MatchError(ContainSubstring("unmarshalling request binding parameters")))
			})

			It("returns an error when secrets cannot be unmarshalled", func() {
				expectedInputParams.CreateBinding.Secrets = "not-json"
				err := action.Execute(expectedInputParams, outputBuffer)
				Expect(err).To(MatchError(ContainSubstring("unmarshalling secrets")))
			})

			It("returns an error when DNS addresses cannot be unmarshalled", func() {
				expectedInputParams.CreateBinding.DNSAddresses = "not-json"
				err := action.Execute(expectedInputParams, outputBuffer)
				Expect(err).To(MatchError(ContainSubstring("unmarshalling DNS addresses")))
			})

			It("returns an generic error when binder returns an error", func() {
				fakeBinder.CreateBindingReturns(serviceadapter.Binding{}, errors.New("something went wrong"))
				err := action.Execute(expectedInputParams, outputBuffer)
				Expect(err).To(BeACLIError(1, "something went wrong"))
			})

			It("returns a BindingAlreadyExistsError when binding already exists", func() {
				fakeBinder.CreateBindingReturns(serviceadapter.Binding{}, serviceadapter.NewBindingAlreadyExistsError(errors.New("something went wrong")))
				err := action.Execute(expectedInputParams, outputBuffer)
				Expect(err).To(BeACLIError(serviceadapter.BindingAlreadyExistsErrorExitCode, "something went wrong"))
			})

			It("returns an AppGuidNotProvidedError when app guid is not provided", func() {
				fakeBinder.CreateBindingReturns(serviceadapter.Binding{}, serviceadapter.NewAppGuidNotProvidedError(errors.New("something went wrong")))
				err := action.Execute(expectedInputParams, outputBuffer)
				Expect(err).To(BeACLIError(serviceadapter.AppGuidNotProvidedErrorExitCode, "something went wrong"))
			})

			It("returns an error when the binding cannot be marshalled", func() {
				fakeBinder.CreateBindingReturns(serviceadapter.Binding{
					Credentials: map[string]interface{}{"a": make(chan bool)},
				},
					nil,
				)

				err := action.Execute(expectedInputParams, outputBuffer)
				Expect(err).To(MatchError(ContainSubstring("error marshalling binding")))
			})
		})
	})
})
