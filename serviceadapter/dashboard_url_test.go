package serviceadapter_test

import (
	"bytes"
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/pivotal-cf/on-demand-services-sdk/bosh"
	"github.com/pivotal-cf/on-demand-services-sdk/serviceadapter"
	"github.com/pivotal-cf/on-demand-services-sdk/serviceadapter/fakes"
)

var _ = Describe("DashboardUrl", func() {
	var (
		fakeDashboardUrlGenerator *fakes.FakeDashboardUrlGenerator
		instanceId                string
		plan                      serviceadapter.Plan
		manifest                  bosh.BoshManifest

		expectedInputParams serviceadapter.InputParams
		action              *serviceadapter.DashboardUrlAction
		outputBuffer        *gbytes.Buffer
	)

	BeforeEach(func() {
		fakeDashboardUrlGenerator = new(fakes.FakeDashboardUrlGenerator)
		instanceId = "my-instance-identifier"
		plan = defaultPlan()
		manifest = defaultManifest()
		outputBuffer = gbytes.NewBuffer()

		expectedInputParams = serviceadapter.InputParams{
			DashboardUrl: serviceadapter.DashboardUrlJSONParams{
				InstanceId: instanceId,
				Plan:       toJson(plan),
				Manifest:   toYaml(manifest),
			},
		}

		action = serviceadapter.NewDashboardUrlAction(fakeDashboardUrlGenerator)
	})

	Describe("IsImplemented", func() {
		It("returns true if implemented", func() {
			Expect(action.IsImplemented()).To(BeTrue())
		})

		It("returns false if not implemented", func() {
			c := serviceadapter.NewDashboardUrlAction(nil)
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
		})

		When("given positional arguments", func() {
			It("can parse positional arguments", func() {
				positionalArgs := []string{
					expectedInputParams.DashboardUrl.InstanceId,
					expectedInputParams.DashboardUrl.Plan,
					expectedInputParams.DashboardUrl.Manifest,
				}

				actualInputParams, err := action.ParseArgs(nil, positionalArgs)
				Expect(err).NotTo(HaveOccurred())
				Expect(actualInputParams).To(Equal(expectedInputParams))
			})

			It("returns an error when required arguments are not passed in", func() {
				_, err := action.ParseArgs(nil, []string{"foo"})
				Expect(err).To(HaveOccurred())
				Expect(err).To(BeAssignableToTypeOf(serviceadapter.MissingArgsError{}))
				Expect(err).To(MatchError(ContainSubstring("<instance-ID> <plan-JSON> <manifest-YAML>")))
			})
		})
	})

	Describe("Execute", func() {
		It("calls the supplied handler passing args through", func() {
			fakeDashboardUrlGenerator.DashboardUrlReturns(serviceadapter.DashboardUrl{DashboardUrl: "gopher://foo"}, nil)

			err := action.Execute(expectedInputParams, outputBuffer)

			Expect(err).NotTo(HaveOccurred())

			Expect(fakeDashboardUrlGenerator.DashboardUrlCallCount()).To(Equal(1))
			actualParams := fakeDashboardUrlGenerator.DashboardUrlArgsForCall(0)

			Expect(actualParams.InstanceID).To(Equal(instanceId))
			Expect(actualParams.Plan).To(Equal(plan))
			Expect(actualParams.Manifest).To(Equal(manifest))

			Expect(outputBuffer).To(gbytes.Say(`{"dashboard_url":"gopher://foo"}`))
		})

		Context("error handling", func() {
			It("returns an error when plan cannot be unmarshalled", func() {
				expectedInputParams.DashboardUrl.Plan = "not-json"
				err := action.Execute(expectedInputParams, outputBuffer)
				Expect(err).To(MatchError(ContainSubstring("unmarshalling service plan")))
			})

			It("returns an error when manifest cannot be unmarshalled", func() {
				expectedInputParams.DashboardUrl.Manifest = "not-yaml"
				err := action.Execute(expectedInputParams, outputBuffer)
				Expect(err).To(MatchError(ContainSubstring("unmarshalling manifest YAML")))
			})

			It("returns an error when plan is invalid", func() {
				expectedInputParams.DashboardUrl.Plan = "{}"
				err := action.Execute(expectedInputParams, outputBuffer)
				Expect(err).To(MatchError(ContainSubstring("validating service plan")))
			})

			It("returns an error when dashboardUrlGenerator returns an error", func() {
				fakeDashboardUrlGenerator.DashboardUrlReturns(serviceadapter.DashboardUrl{}, errors.New("something went wrong"))
				err := action.Execute(expectedInputParams, outputBuffer)
				Expect(err).To(BeACLIError(1, "something went wrong"))
			})

			It("returns an error when dashboardUrlGenerator returns an unmarshalable struct", func() {
				fakeWriter := new(FakeWriter)
				fakeDashboardUrlGenerator.DashboardUrlReturns(serviceadapter.DashboardUrl{}, nil)

				err := action.Execute(expectedInputParams, fakeWriter)
				Expect(err).To(MatchError(ContainSubstring("marshalling dashboardUrl")))
			})
		})
	})
})
