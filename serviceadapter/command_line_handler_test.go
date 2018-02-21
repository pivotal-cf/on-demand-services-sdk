package serviceadapter_test

import (
	"errors"
	"io"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"

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

		outputWriter io.Writer
		errorWriter  io.Writer
	)

	BeforeEach(func() {
		fakeManifestGenerator = new(fakes.FakeManifestGenerator)
		fakeBinder = new(fakes.FakeBinder)
		fakeDashboardUrlGenerator = new(fakes.FakeDashboardUrlGenerator)
		outputWriter = gbytes.NewBuffer()
		errorWriter = gbytes.NewBuffer()

		handler = serviceadapter.CommandLineHandler{
			ManifestGenerator:     fakeManifestGenerator,
			Binder:                fakeBinder,
			DashboardURLGenerator: fakeDashboardUrlGenerator,
		}

		serviceDeployment = serviceadapter.ServiceDeployment{
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

		args = serviceadapter.RequestParameters{"key": "foo", "bar": "baz"}

		plan = serviceadapter.Plan{
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

		previousPlan = serviceadapter.Plan{
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

		previousManifest = bosh.BoshManifest{Name: "another-deployment-name",
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

		serviceDeploymentJSON = toJson(serviceDeployment)
		planJSON = toJson(plan)
		previousPlanJSON = toJson(previousPlan)
		argsJSON = toJson(args)
		previousManifestYAML = toYaml(previousManifest)
	})

	It("when no arguments passed returns error", func() {
		err := handler.Handle([]string{commandName}, outputWriter, errorWriter)
		assertCLIHandlerErr(err, serviceadapter.ErrorExitCode, "following commands are supported")
	})

	Describe("Generate Manifest", func() {
		It("calls the supplied handler passing args through", func() {
			manifest := bosh.BoshManifest{Name: "bill"}
			fakeManifestGenerator.GenerateManifestReturns(manifest, nil)

			err := handler.Handle([]string{
				commandName, "generate-manifest", serviceDeploymentJSON, planJSON, argsJSON, previousManifestYAML, previousPlanJSON,
			}, outputWriter, errorWriter)
			Expect(err).NotTo(HaveOccurred())

			Expect(fakeManifestGenerator.GenerateManifestCallCount()).To(Equal(1))
			actualServiceDeployment, actualPlan, actualRequestParams, actualPreviousManifest, actualPreviousPlan :=
				fakeManifestGenerator.GenerateManifestArgsForCall(0)

			Expect(actualServiceDeployment).To(Equal(serviceDeployment))
			Expect(actualPlan).To(Equal(plan))
			Expect(actualRequestParams).To(Equal(args))
			Expect(actualPreviousManifest).To(Equal(&previousManifest))
			Expect(actualPreviousPlan).To(Equal(&previousPlan))

			Expect(outputWriter).To(gbytes.Say("bill"))
		})

		It("returns a not-implemented error when there is no generate manifest handler", func() {
			handler.ManifestGenerator = nil
			err := handler.Handle([]string{
				commandName,
				"generate-manifest",
				serviceDeploymentJSON,
				planJSON,
				argsJSON,
				previousManifestYAML,
				previousPlanJSON,
			}, outputWriter, errorWriter)
			assertCLIHandlerErr(err, serviceadapter.NotImplementedExitCode, "manifest generator not implemented")
		})

		It("returns a missing args error when previous plan json is missing", func() {
			err := handler.Handle([]string{
				commandName, "generate-manifest", serviceDeploymentJSON, planJSON, argsJSON, previousManifestYAML,
			}, outputWriter, errorWriter)
			assertCLIHandlerErr(err, serviceadapter.ErrorExitCode, `Missing arguments for generate-manifest. Usage:`)
		})

		It("returns a error if service deployment JSON is corrupt", func() {
			serviceDeploymentJSON += "asdf"
			err := handler.Handle([]string{
				commandName, "generate-manifest", serviceDeploymentJSON, planJSON, argsJSON, previousManifestYAML, previousPlanJSON,
			}, outputWriter, errorWriter)
			Expect(err).To(MatchError(ContainSubstring("unmarshalling service deployment")))
		})

		It("returns a error if service deployment JSON is invalid", func() {
			serviceDeployment.Releases = nil
			serviceDeploymentJSON = toJson(serviceDeployment)
			err := handler.Handle([]string{
				commandName, "generate-manifest", serviceDeploymentJSON, planJSON, argsJSON, previousManifestYAML, previousPlanJSON,
			}, outputWriter, errorWriter)
			Expect(err).To(MatchError(ContainSubstring("validating service deployment")))
		})

		It("returns a error if service plan JSON is corrupt", func() {
			planJSON += "asdf"
			err := handler.Handle([]string{
				commandName, "generate-manifest", serviceDeploymentJSON, planJSON, argsJSON, previousManifestYAML, previousPlanJSON,
			}, outputWriter, errorWriter)
			Expect(err).To(MatchError(ContainSubstring("unmarshalling service plan")))
		})

		It("returns a error if service plan JSON is invalid", func() {
			plan.InstanceGroups = nil
			planJSON = toJson(plan)
			err := handler.Handle([]string{
				commandName, "generate-manifest", serviceDeploymentJSON, planJSON, argsJSON, previousManifestYAML, previousPlanJSON,
			}, outputWriter, errorWriter)
			Expect(err).To(MatchError(ContainSubstring("validating service plan")))
		})

		It("returns a error if request params JSON is corrupt", func() {
			argsJSON += "asdf"
			err := handler.Handle([]string{
				commandName, "generate-manifest", serviceDeploymentJSON, planJSON, argsJSON, previousManifestYAML, previousPlanJSON,
			}, outputWriter, errorWriter)
			Expect(err).To(MatchError(ContainSubstring("unmarshalling requestParams")))
		})

		It("returns a error if previous manifest YAML is corrupt", func() {
			previousManifestYAML = planJSON
			err := handler.Handle([]string{
				commandName, "generate-manifest", serviceDeploymentJSON, planJSON, argsJSON, previousManifestYAML, previousPlanJSON,
			}, outputWriter, errorWriter)
			Expect(err).To(MatchError(ContainSubstring("unmarshalling previous manifest")))
		})

		It("returns a error if previous service plan JSON is corrupt", func() {
			previousPlanJSON += "asdf"
			err := handler.Handle([]string{
				commandName, "generate-manifest", serviceDeploymentJSON, planJSON, argsJSON, previousManifestYAML, previousPlanJSON,
			}, outputWriter, errorWriter)
			Expect(err).To(MatchError(ContainSubstring("unmarshalling previous service plan")))
		})

		It("returns a error if previous service plan JSON is invalid", func() {
			previousPlan.InstanceGroups = nil
			previousPlanJSON = toJson(previousPlan)
			err := handler.Handle([]string{
				commandName, "generate-manifest", serviceDeploymentJSON, planJSON, argsJSON, previousManifestYAML, previousPlanJSON,
			}, outputWriter, errorWriter)
			Expect(err).To(MatchError(ContainSubstring("validating previous service plan")))
		})

		It("returns an error when the manifest cannot be generated", func() {
			fakeManifestGenerator.GenerateManifestReturns(bosh.BoshManifest{}, errors.New("oops"))
			err := handler.Handle([]string{
				commandName, "generate-manifest", serviceDeploymentJSON, planJSON, argsJSON, previousManifestYAML, previousPlanJSON,
			}, outputWriter, errorWriter)

			assertCLIHandlerErr(err, serviceadapter.ErrorExitCode, "oops")
			Expect(outputWriter).To(gbytes.Say("oops"))
		})
	})
})

func assertCLIHandlerErr(err error, exitCode int, message string) {
	Expect(err).To(BeAssignableToTypeOf(serviceadapter.CLIHandlerError{}))
	Expect(err).To(MatchError(ContainSubstring(message)))
	actualErr := err.(serviceadapter.CLIHandlerError)
	Expect(actualErr.ExitCode).To(Equal(exitCode))
}
