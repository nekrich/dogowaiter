package dogowaiterdocker

import (
	"context"
	"dogowaiter/internal/dogowaiterdocker/dogowaiterdockercurrent"
	"dogowaiter/internal/dogowaiteroptions"
	"fmt"
	"log/slog"
	"slices"
	"strings"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
)

const (
	dockerComposeServiceLabel = "com.docker.compose.service"
	dockerComposeProjectLabel = "com.docker.compose.project"
	dogowaiterEnabledLabel    = "dogowaiter.enabled"
	dogowaiterDependsOnLabel  = "dogowaiter.depends_on"
)

// DockerContainerResolveResult holds the container IDs and names resolved for one dependency.
type DockerContainerResolveResult struct {
	DependencyName string
	ContainerIDs   []string
	ContainerNames []string
}

func (lhs DockerContainerResolveResult) Equal(rhs DockerContainerResolveResult) bool {
	return lhs.DependencyName == rhs.DependencyName && slices.Equal(lhs.ContainerIDs, rhs.ContainerIDs) && slices.Equal(lhs.ContainerNames, rhs.ContainerNames)
}

type DogowaiterDockerContainerResolverInterface interface {
	FindRequiredDockerContainers(ctx context.Context, cli *client.Client, deps []dogowaiteroptions.DogowaiterConfigurationDependency, logger *slog.Logger) ([]DockerContainerResolveResult, error)
}

type DogowaiterDockerContainerResolver struct{}

func (resolver DogowaiterDockerContainerResolver) FindDependentDockerContainersInStack(ctx context.Context, cli *client.Client, logger *slog.Logger) ([]dogowaiteroptions.DogowaiterConfigurationDependency, error) {

	info := dogowaiterdockercurrent.DogowaiterDockerInfo{
		ContainerInspecter:  cli,
		ContainerIDProvider: dogowaiterdockercurrent.DogowaiterDockerCurrentContainerIDProvider{},
		Logger:              logger,
	}

	stackName := info.GetStackName(ctx)
	logger.Info("stack name", "stackName", stackName)
	if stackName == "" {
		return nil, nil
	}

	// Docker API treats multiple "label" filter values as OR, not AND. So we filter by
	// project only here, then require all labels in code.
	filters := filters.NewArgs()
	filters.Add("label", fmt.Sprintf("%s=%s", dockerComposeProjectLabel, stackName))
	filters.Add("label", fmt.Sprintf("%s=true", dogowaiterEnabledLabel))
	filters.Add("label", dogowaiterDependsOnLabel)
	logger.Debug("filters", "filters", filters)

	containers, err := cli.ContainerList(ctx, container.ListOptions{All: true, Filters: filters})
	if err != nil {
		logger.Error("failed to list docker containers for stack", "error", err)
		return nil, err
	}
	logger.Debug("containers found", "count", len(containers))

	dependencies := []dogowaiteroptions.DogowaiterConfigurationDependency{}
	for _, container := range containers {
		if _, ok := container.Labels[dogowaiterEnabledLabel]; !ok {
			logger.Debug("container is not enabled", "container", container.ID)
			continue
		}
		if label, ok := container.Labels[dockerComposeProjectLabel]; !ok || label != stackName {
			logger.Debug("container is not in stack", "container", container.ID, "stack", stackName)
			continue
		}
		if label, ok := container.Labels[dogowaiterDependsOnLabel]; ok {
			containerDependencies := dogowaiteroptions.DogowaiterConfigurationDependenciesFromString(label)
			logger.Debug("container dependencies", "container", container.ID, "dependencies", containerDependencies)
			for _, dependency := range containerDependencies {
				if slices.ContainsFunc(dependencies, func(lhs dogowaiteroptions.DogowaiterConfigurationDependency) bool {
					return lhs.Equal(dependency)
				}) {
					logger.Debug("dependency already exists", "dependency", dependency)
					continue
				}
				dependencies = append(dependencies, dependency)
			}
		} else {
			logger.Debug("container has no depends_on label", "container", container.ID)
		}
	}
	return dependencies, nil
}

// FindDockerContainers lists running containers and matches each dependency by stack (optional) and service name.
func (resolver DogowaiterDockerContainerResolver) FindRequiredDockerContainers(ctx context.Context, cli *client.Client, dependencies []dogowaiteroptions.DogowaiterConfigurationDependency, logger *slog.Logger) ([]DockerContainerResolveResult, error) {
	logger.Debug("finding required docker containers", "dependencies", len(dependencies))
	containers, err := cli.ContainerList(ctx, container.ListOptions{})
	if err != nil {
		logger.Error("failed to list docker containers", "error", err)
		return nil, err
	}
	logger.Debug("docker containers found", "containerCount", len(containers))

	results := make([]DockerContainerResolveResult, 0, len(dependencies))
	for _, dependency := range dependencies {
		var containerIDs, containerNames []string
		for _, container := range containers {
			if matchContainer(dependency, container) {
				containerIDs = append(containerIDs, container.ID)
				name := strings.TrimPrefix(container.Names[0], "/")
				containerNames = append(containerNames, name)
			}
		}
		dependencyResult := DockerContainerResolveResult{
			DependencyName: dependency.Name,
			ContainerIDs:   containerIDs,
			ContainerNames: containerNames,
		}
		if slices.ContainsFunc(results, func(result DockerContainerResolveResult) bool {
			return result.Equal(dependencyResult)
		}) {
			logger.Debug("dependency result already exists", "dependencyResult", dependencyResult)
			continue
		}
		results = append(results, dependencyResult)
	}
	return results, nil
}

// matchContainer returns true if the container c matches the dependency (stack and service name).
func matchContainer(dependency dogowaiteroptions.DogowaiterConfigurationDependency, container container.Summary) bool {
	if dependency.Stack != "" {
		if label, ok := container.Labels[dockerComposeProjectLabel]; !ok || label != dependency.Stack {
			return false
		}
	}
	if label, ok := container.Labels[dockerComposeServiceLabel]; ok && label == dependency.Name {
		return true
	}
	for _, n := range container.Names {
		name := strings.TrimPrefix(n, "/")
		if name == dependency.Name || strings.Contains(name, dependency.Name) {
			return true
		}
	}
	return false
}

// GetWatchedContainerIDs returns the set of all container IDs from results.
func GetWatchedContainerIDs(results []DockerContainerResolveResult) map[string]struct{} {
	m := make(map[string]struct{})
	for _, r := range results {
		for _, id := range r.ContainerIDs {
			m[id] = struct{}{}
		}
	}
	return m
}
