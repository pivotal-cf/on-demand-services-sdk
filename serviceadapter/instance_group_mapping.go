package serviceadapter

import (
	"fmt"
	"strings"

	"github.com/pivotal-cf/on-demand-services-sdk/bosh"
)

func GenerateInstanceGroupsWithNoProperties(
	instanceGroups []InstanceGroup,
	serviceReleases ServiceReleases,
	stemcell string,
	deploymentInstanceGroupsToJobs map[string][]string,
) ([]bosh.InstanceGroup, error) {

	if len(instanceGroups) == 0 {
		return nil, fmt.Errorf("no instance groups provided")
	}

	boshInstanceGroups := []bosh.InstanceGroup{}
	for _, instanceGroup := range instanceGroups {
		if _, ok := deploymentInstanceGroupsToJobs[instanceGroup.Name]; !ok {
			continue
		}

		networks := []bosh.Network{}

		for _, network := range instanceGroup.Networks {
			networks = append(networks, bosh.Network{Name: network})
		}

		boshJobs, err := generateJobsForInstanceGroup(instanceGroup.Name, deploymentInstanceGroupsToJobs, serviceReleases)
		if err != nil {
			return nil, err
		}
		boshInstanceGroup := bosh.InstanceGroup{
			Name:               instanceGroup.Name,
			Instances:          instanceGroup.Instances,
			Stemcell:           stemcell,
			VMType:             instanceGroup.VMType,
			VMExtensions:       instanceGroup.VMExtensions,
			PersistentDiskType: instanceGroup.PersistentDiskType,
			AZs:                instanceGroup.AZs,
			Networks:           networks,
			Jobs:               boshJobs,
			Lifecycle:          instanceGroup.Lifecycle,
		}
		boshInstanceGroups = append(boshInstanceGroups, boshInstanceGroup)
	}
	return boshInstanceGroups, nil
}

func findReleaseForJob(jobName string, releases ServiceReleases) (string, error) {
	releasesThatMentionJob := []string{}
	for _, release := range releases {
		for _, job := range release.Jobs {
			if job == jobName {
				releasesThatMentionJob = append(releasesThatMentionJob, release.Name)
			}
		}
	}

	if len(releasesThatMentionJob) == 0 {
		return "", fmt.Errorf("job '%s' not provided", jobName)
	}

	if len(releasesThatMentionJob) > 1 {
		return "", fmt.Errorf("job '%s' provided %d times, by %s", jobName, len(releasesThatMentionJob), strings.Join(releasesThatMentionJob, ", "))
	}

	return releasesThatMentionJob[0], nil
}
func generateJobsForInstanceGroup(instanceGroupName string, deploymentInstanceGroupsToJobs map[string][]string, serviceReleases ServiceReleases) ([]bosh.Job, error) {
	boshJobs := []bosh.Job{}
	for _, job := range deploymentInstanceGroupsToJobs[instanceGroupName] {
		release, err := findReleaseForJob(job, serviceReleases)
		if err != nil {
			return nil, err
		}

		boshJobs = append(boshJobs, bosh.Job{Name: job, Release: release})
	}
	return boshJobs, nil
}
