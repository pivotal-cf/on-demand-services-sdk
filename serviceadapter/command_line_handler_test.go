package serviceadapter_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"

	"gopkg.in/yaml.v2"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/on-demand-service-broker-sdk/bosh"
	"github.com/pivotal-cf/on-demand-service-broker-sdk/serviceadapter"
	"github.com/pivotal-cf/on-demand-service-broker-sdk/serviceadapter/fake_service_adapter"
)

var _ = Describe("Command line handler", func() {
	var manifestGenerator *fake_service_adapter.FakeManifestGenerator
	var binder *fake_service_adapter.FakeBinder
	var dashboardUrlGenerator *fake_service_adapter.FakeDashboardUrlGenerator
	var args []string
	var outputBuffer *bytes.Buffer
	var logBuffer *bytes.Buffer
	var exitCode int

	var (
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
				Name:           "example-server",
				VMType:         "small",
				PersistentDisk: "ten",
				Networks:       []string{"example-network"},
				AZs:            []string{"example-az"},
				Instances:      1,
				Lifecycle:      "errand",
			}},
			Properties: serviceadapter.Properties{"example": "property"},
		}
		expectedAribtaryParams = map[string]interface{}{"key": "foo", "bar": "baz"}

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
			}}

		expectedPreviousPlan = serviceadapter.Plan{
			InstanceGroups: []serviceadapter.InstanceGroup{{
				Name:           "another-example-server",
				VMType:         "small",
				PersistentDisk: "ten",
				Networks:       []string{"example-network"},
				AZs:            []string{"example-az"},
				Instances:      1,
				Lifecycle:      "errand",
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
		expectedBoshVMs   = bosh.BoshVMs{"kafka": []string{"a", "b"}}
		expectedManifest  = expectedPreviousManifest

		expectedResultantBinding = serviceadapter.Binding{
			RouteServiceURL: "a route",
			SyslogDrainURL:  "a url",
			Credentials: map[string]interface{}{
				"binding": "this binds",
			},
		}
	)

	BeforeEach(func() {
		args = []string{}
		outputBuffer = bytes.NewBuffer([]byte{})
		logBuffer = bytes.NewBuffer([]byte{})

		serviceadapter.OutputWriter = io.MultiWriter(outputBuffer, GinkgoWriter)

		manifestGenerator = &fake_service_adapter.FakeManifestGenerator{}
		binder = &fake_service_adapter.FakeBinder{}
		dashboardUrlGenerator = &fake_service_adapter.FakeDashboardUrlGenerator{}
		exitCode = 0
		serviceadapter.Exiter = func(code int) { exitCode = code }
	})

	AfterEach(func() {
		serviceadapter.OutputWriter = os.Stdout
		serviceadapter.Exiter = os.Exit
	})
	var When = Context

	When("interfaces are fully implemented", func() {

		JustBeforeEach(func() {
			serviceadapter.HandleCommandLineInvocation(args, manifestGenerator, binder, dashboardUrlGenerator, log.New(io.MultiWriter(logBuffer, GinkgoWriter), "[on-demand-service-adapter-test] ", log.LstdFlags))
		})

		Context("generating a manifest", func() {
			BeforeEach(func() {
				args = []string{"command-name", "generate-manifest", toJson(expectedServiceDeployment), toJson(expectedCurrentPlan), toJson(expectedAribtaryParams), "", "null"}
				manifestGenerator.GenerateManifestReturns(expectedResultantBoshManifest, nil)
			})

			var (
				actualServiceDeployment serviceadapter.ServiceDeployment
				acutalCurrentPlan       serviceadapter.Plan
				acutalAribtaryParams    map[string]interface{}
				actualPreviousManifest  *bosh.BoshManifest
				actualPreviousPlan      *serviceadapter.Plan
			)
			JustBeforeEach(func() {
				actualServiceDeployment, acutalCurrentPlan, acutalAribtaryParams, actualPreviousManifest, actualPreviousPlan = manifestGenerator.GenerateManifestArgsForCall(0)
			})
			It("only invokes generate manifest", func() {
				Expect(binder.CreateBindingCallCount()).To(BeZero())
				Expect(binder.DeleteBindingCallCount()).To(BeZero())
				Expect(dashboardUrlGenerator.DashboardUrlCallCount()).To(BeZero())
				Expect(manifestGenerator.GenerateManifestCallCount()).To(Equal(1))
			})

			It("deserialises the service deployment", func() {
				Expect(actualServiceDeployment).To(Equal(expectedServiceDeployment))
			})

			It("deserialises the current plan", func() {
				Expect(acutalCurrentPlan).To(Equal(expectedCurrentPlan))
			})

			It("deserialises the aribitary params", func() {
				Expect(acutalAribtaryParams).To(Equal(expectedAribtaryParams))
			})

			It("deserialises the manfiest as nil", func() {
				Expect(actualPreviousManifest).To(BeNil())
			})

			It("deserialises the previous plan as nil", func() {
				Expect(actualPreviousPlan).To(BeNil())
			})

			It("serialzies the manifest as yaml", func() {
				Expect(outputBuffer.String()).To(Equal(toYaml(expectedResultantBoshManifest)))
			})

			Context("when optional paramters are passed in", func() {
				BeforeEach(func() {
					args = []string{"command-name", "generate-manifest", toJson(expectedServiceDeployment), toJson(expectedCurrentPlan), toJson(expectedAribtaryParams), toYaml(expectedPreviousManifest), toJson(expectedPreviousPlan)}
				})

				It("deserialises the manfiest from params", func() {
					Expect(actualPreviousManifest).To(Equal(&expectedPreviousManifest))
				})

				It("deserialises the previous plan from params", func() {
					Expect(actualPreviousPlan).To(Equal(&expectedPreviousPlan))
				})
			})

			Context("error generating a manifest", func() {
				BeforeEach(func() {
					manifestGenerator.GenerateManifestReturns(bosh.BoshManifest{}, fmt.Errorf("not valid"))
				})
				It("Fails and logs", func() {
					Expect(exitCode).To(Equal(1))
					Expect(logBuffer).To(ContainSubstring("not valid"))
				})
			})
		})

		Context("binding", func() {
			var (
				actualBindingId     string
				actualBoshVMs       bosh.BoshVMs
				actualBoshManifest  bosh.BoshManifest
				actualBindingParams map[string]interface{}
			)
			JustBeforeEach(func() {
				actualBindingId, actualBoshVMs, actualBoshManifest, actualBindingParams = binder.CreateBindingArgsForCall(0)
			})

			BeforeEach(func() {
				args = []string{"command-name", "create-binding", expectedBindingID, toJson(expectedBoshVMs), toYaml(expectedManifest), toJson(expectedAribtaryParams)}
				binder.CreateBindingReturns(expectedResultantBinding, nil)
			})

			It("only invokes create binding", func() {
				Expect(binder.CreateBindingCallCount()).To(Equal(1))
				Expect(binder.DeleteBindingCallCount()).To(BeZero())
				Expect(dashboardUrlGenerator.DashboardUrlCallCount()).To(BeZero())
				Expect(manifestGenerator.GenerateManifestCallCount()).To(BeZero())
			})

			It("reads the binding id", func() {
				Expect(actualBindingId).To(Equal(expectedBindingID))
			})

			It("deserializes the bosh vms", func() {
				Expect(actualBoshVMs).To(Equal(expectedBoshVMs))
			})
			It("deserializes the manifest", func() {
				Expect(actualBoshManifest).To(Equal(expectedManifest))
			})
			It("deserializes the aribitary params", func() {
				Expect(actualBindingParams).To(Equal(expectedAribtaryParams))
			})

			It("serialzies binding result as json", func() {
				Expect(outputBuffer.String()).To(MatchJSON(toJson(expectedResultantBinding)))
			})

			Context("binding fails", func() {
				Context("binding already exists", func() {
					BeforeEach(func() {
						binder.CreateBindingReturns(serviceadapter.Binding{}, serviceadapter.NewBindingAlreadyExistsError(fmt.Errorf("binding foo already exists")))
					})
					It("Fails and logs", func() {
						Expect(exitCode).To(Equal(49))
						Expect(logBuffer).To(ContainSubstring("binding foo already exists"))
					})
				})

				Context("app_guid isn't provided", func() {
					BeforeEach(func() {
						binder.CreateBindingReturns(serviceadapter.Binding{}, serviceadapter.NewAppGuidNotProvidedError(fmt.Errorf("no app_guid provided")))
					})

					It("Fails and logs", func() {
						Expect(exitCode).To(Equal(42))
						Expect(logBuffer).To(ContainSubstring("no app_guid provided"))
					})

				})

				Context("internal error", func() {
					BeforeEach(func() {
						binder.CreateBindingReturns(serviceadapter.Binding{}, fmt.Errorf("not valid"))
					})
					It("Fails and logs", func() {
						Expect(exitCode).To(Equal(1))
						Expect(logBuffer).To(ContainSubstring("not valid"))
					})
				})
			})
		})
		Context("unbinding", func() {
			var (
				actualBindingId    string
				actualBoshVMs      bosh.BoshVMs
				actualBoshManifest bosh.BoshManifest
			)
			JustBeforeEach(func() {
				actualBindingId, actualBoshVMs, actualBoshManifest = binder.DeleteBindingArgsForCall(0)
			})

			BeforeEach(func() {
				args = []string{"command-name", "delete-binding", expectedBindingID, toJson(expectedBoshVMs), toYaml(expectedManifest), toJson(expectedAribtaryParams)}
				binder.DeleteBindingReturns(nil)
			})

			It("only invokes delete binding", func() {
				Expect(binder.CreateBindingCallCount()).To(BeZero())
				Expect(binder.DeleteBindingCallCount()).To(Equal(1))
				Expect(dashboardUrlGenerator.DashboardUrlCallCount()).To(BeZero())
				Expect(manifestGenerator.GenerateManifestCallCount()).To(BeZero())
			})

			It("reads the binding id", func() {
				Expect(actualBindingId).To(Equal(expectedBindingID))
			})

			It("deserializes the bosh vms", func() {
				Expect(actualBoshVMs).To(Equal(expectedBoshVMs))
			})
			It("deserializes the manifest", func() {
				Expect(actualBoshManifest).To(Equal(expectedManifest))
			})

			Context("binding fails", func() {
				BeforeEach(func() {
					binder.DeleteBindingReturns(fmt.Errorf("not valid"))
				})
				It("Fails and logs", func() {
					Expect(exitCode).To(Equal(1))
					Expect(logBuffer).To(ContainSubstring("not valid"))
				})
			})
		})
		Context("dashboard-url", func() {
			var (
				actualInstanceID string
				actualPlan       serviceadapter.Plan
				actualManifest   bosh.BoshManifest
			)
			JustBeforeEach(func() {
				actualInstanceID, actualPlan, actualManifest = dashboardUrlGenerator.DashboardUrlArgsForCall(0)
			})

			BeforeEach(func() {
				args = []string{"command-name", "dashboard-url", "instance-identifier", toJson(expectedPlan), toYaml(expectedManifest)}
				dashboardUrlGenerator.DashboardUrlReturns(serviceadapter.DashboardUrl{DashboardUrl: "http://dashboard.com"}, nil)
			})
			It("only invokes dashboard url generator", func() {
				Expect(binder.CreateBindingCallCount()).To(BeZero())
				Expect(binder.DeleteBindingCallCount()).To(BeZero())
				Expect(manifestGenerator.GenerateManifestCallCount()).To(BeZero())
				Expect(dashboardUrlGenerator.DashboardUrlCallCount()).To(Equal(1))
			})

			It("passes through the instance id", func() {
				Expect(actualInstanceID).To(Equal("instance-identifier"))
			})

			It("deserializes the plan", func() {
				Expect(actualPlan).To(Equal(expectedPlan))
			})

			It("deserializes the manifest", func() {
				Expect(actualManifest).To(Equal(expectedManifest))
			})

			It("exits with 0", func() {
				Expect(exitCode).To(BeZero())
			})
			It("returns the dashboard URL", func() {
				Expect(outputBuffer.Bytes()).To(MatchJSON(`{ "dashboard_url": "http://dashboard.com"}`))
			})

			Context("when it fails", func() {
				BeforeEach(func() {
					dashboardUrlGenerator.DashboardUrlReturns(serviceadapter.DashboardUrl{}, fmt.Errorf("cant do it"))
				})
				It("exits with 1", func() {
					Expect(exitCode).To(Equal(1))
				})
				It("logs the error", func() {
					Expect(logBuffer).To(ContainSubstring("cant do it"))
				})
			})
		})
	})

	Context("supporting parts of the interface", func() {
		When("manifest generator isn't implemented", func() {
			BeforeEach(func() {
				args = []string{"command-name", "generate-manifest", toJson(expectedServiceDeployment), toJson(expectedCurrentPlan), toJson(expectedAribtaryParams), "", "null"}
				serviceadapter.HandleCommandLineInvocation(args, nil, binder, dashboardUrlGenerator, log.New(io.MultiWriter(logBuffer, GinkgoWriter), "[on-demand-service-adapter-test] ", log.LstdFlags))
			})
			It("exits with 10", func() {
				Expect(exitCode).To(Equal(10))
			})
		})

		When("service binder isn't implemented", func() {
			var command string
			JustBeforeEach(func() {
				args = []string{"command-name", command, toJson(expectedServiceDeployment), toJson(expectedCurrentPlan), toJson(expectedAribtaryParams), "", "null"}
				serviceadapter.HandleCommandLineInvocation(args, manifestGenerator, nil, dashboardUrlGenerator, log.New(io.MultiWriter(logBuffer, GinkgoWriter), "[on-demand-service-adapter-test] ", log.LstdFlags))
			})
			Context("create-binding", func() {
				BeforeEach(func() {
					command = "create-binding"
				})
				It("exits with 10", func() {
					Expect(exitCode).To(Equal(10))
				})
			})
			Context("delete-binding", func() {
				BeforeEach(func() {
					command = "delete-binding"
				})
				It("exits with 10", func() {
					Expect(exitCode).To(Equal(10))
				})
			})
		})

		When("dashboard url generator isn't implemented", func() {
			BeforeEach(func() {
				args = []string{"command-name", "dashboard-url", "id", toJson(expectedCurrentPlan), "null"}
				serviceadapter.HandleCommandLineInvocation(args, manifestGenerator, binder, nil, log.New(io.MultiWriter(logBuffer, GinkgoWriter), "[on-demand-service-adapter-test] ", log.LstdFlags))
			})
			It("exits with 10", func() {
				Expect(exitCode).To(Equal(10))
			})
		})
	})
})

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
