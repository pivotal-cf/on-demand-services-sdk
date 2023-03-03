package serviceadapter_test

import (
	"bytes"
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/pivotal-cf/on-demand-services-sdk/bosh"
	"github.com/pivotal-cf/on-demand-services-sdk/serviceadapter"
	"github.com/pivotal-cf/on-demand-services-sdk/serviceadapter/fakes"
)

var _ = Describe("DeleteBinding", func() {
	var (
		fakeBinder    *fakes.FakeBinder
		bindingId     string
		boshVMs       bosh.BoshVMs
		requestParams serviceadapter.RequestParameters
		secrets       serviceadapter.ManifestSecrets
		dnsAddresses  serviceadapter.DNSAddresses
		manifest      bosh.BoshManifest

		expectedInputParams serviceadapter.InputParams
		action              *serviceadapter.DeleteBindingAction
		outputBuffer        *gbytes.Buffer
	)

	BeforeEach(func() {
		fakeBinder = new(fakes.FakeBinder)
		bindingId = "binding-id"
		boshVMs = bosh.BoshVMs{}
		requestParams = defaultRequestParams()
		secrets = defaultSecretParams()
		dnsAddresses = defaultDNSParams()
		manifest = defaultManifest()
		outputBuffer = gbytes.NewBuffer()

		expectedInputParams = serviceadapter.InputParams{
			DeleteBinding: serviceadapter.DeleteBindingJSONParams{
				BindingId:         bindingId,
				BoshVms:           toJson(boshVMs),
				Manifest:          toYaml(manifest),
				RequestParameters: toJson(requestParams),
			},
		}

		action = serviceadapter.NewDeleteBindingAction(fakeBinder)
	})

	Describe("IsImplemented", func() {
		It("returns true if implemented", func() {
			Expect(action.IsImplemented()).To(BeTrue())
		})

		It("returns false if not implemented", func() {
			c := serviceadapter.NewDeleteBindingAction(nil)
			Expect(c.IsImplemented()).To(BeFalse())
		})
	})

	Describe("ParseArgs", func() {
		When("giving arguments in stdin", func() {
			It("can parse arguments from stdin", func() {
				expectedInputParams.DeleteBinding.Secrets = toJson(secrets)
				expectedInputParams.DeleteBinding.DNSAddresses = toJson(dnsAddresses)
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
					expectedInputParams.DeleteBinding.BindingId,
					expectedInputParams.DeleteBinding.BoshVms,
					expectedInputParams.DeleteBinding.Manifest,
					expectedInputParams.DeleteBinding.RequestParameters,
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
			fakeBinder.DeleteBindingReturns(nil)

			expectedInputParams.DeleteBinding.Secrets = toJson(secrets)
			expectedInputParams.DeleteBinding.DNSAddresses = toJson(dnsAddresses)
			err := action.Execute(expectedInputParams, outputBuffer)

			Expect(err).NotTo(HaveOccurred())

			Expect(fakeBinder.DeleteBindingCallCount()).To(Equal(1))
			params := fakeBinder.DeleteBindingArgsForCall(0)

			Expect(params.BindingID).To(Equal(bindingId))
			Expect(params.DeploymentTopology).To(Equal(boshVMs))
			Expect(params.Manifest).To(Equal(manifest))
			Expect(params.RequestParams).To(Equal(requestParams))
			Expect(params.Secrets).To(Equal(secrets))
			Expect(params.DNSAddresses).To(Equal(dnsAddresses))
		})

		Context("error handling", func() {
			It("returns an error when bosh VMs cannot be unmarshalled", func() {
				expectedInputParams.DeleteBinding.BoshVms = "not-json"
				err := action.Execute(expectedInputParams, outputBuffer)
				Expect(err).To(MatchError(ContainSubstring("unmarshalling BOSH VMs")))
			})

			It("returns an error when manifest cannot be unmarshalled", func() {
				expectedInputParams.DeleteBinding.Manifest = "not-yaml"
				err := action.Execute(expectedInputParams, outputBuffer)
				Expect(err).To(MatchError(ContainSubstring("unmarshalling manifest YAML")))
			})

			It("returns an error when request params cannot be unmarshalled", func() {
				expectedInputParams.DeleteBinding.RequestParameters = "not-json"
				err := action.Execute(expectedInputParams, outputBuffer)
				Expect(err).To(MatchError(ContainSubstring("unmarshalling request binding parameters")))
			})

			It("returns an error when secrets cannot be unmarshalled", func() {
				expectedInputParams.DeleteBinding.Secrets = "not-json"
				err := action.Execute(expectedInputParams, outputBuffer)
				Expect(err).To(MatchError(ContainSubstring("unmarshalling secrets")))
			})

			It("returns an error when DNS addresses cannot be unmarshalled", func() {
				expectedInputParams.DeleteBinding.DNSAddresses = "not-json"
				err := action.Execute(expectedInputParams, outputBuffer)
				Expect(err).To(MatchError(ContainSubstring("unmarshalling DNS addresses")))
			})

			It("returns an error when binder returns an error", func() {
				fakeBinder.DeleteBindingReturns(errors.New("something went wrong"))
				err := action.Execute(expectedInputParams, outputBuffer)
				Expect(err).To(BeACLIError(1, "something went wrong"))
			})

			It("returns a BindingNotFoundError when binding not found", func() {
				fakeBinder.DeleteBindingReturns(serviceadapter.NewBindingNotFoundError(errors.New("something went wrong")))
				err := action.Execute(expectedInputParams, outputBuffer)
				Expect(err).To(BeACLIError(serviceadapter.BindingNotFoundErrorExitCode, "something went wrong"))
			})
		})
	})
})
