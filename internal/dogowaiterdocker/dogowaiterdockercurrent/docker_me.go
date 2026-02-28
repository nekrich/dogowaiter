package dogowaiterdockercurrent

import (
	"context"
	"log/slog"

	"github.com/docker/docker/api/types/container"
)

const (
	dockerComposeProjectLabel = "com.docker.compose.project"
)

// DogowaiterDockerInfo provides information about the current container.
type DogowaiterDockerInfo struct {
	ContainerInspecter  DogowaiterDockerContainerInspector
	ContainerIDProvider DogowaiterDockerContainerIDProvider
	Logger              *slog.Logger
	container           *container.InspectResponse
	stackName           string
}

func (info *DogowaiterDockerInfo) getContainerID() string {
	return info.ContainerIDProvider.ContainerID()
}

func (info *DogowaiterDockerInfo) getCurrentContainer(ctx context.Context) *container.InspectResponse {
	if info.container != nil {
		return info.container
	}
	if info.ContainerInspecter == nil {
		return nil
	}
	logger := info.Logger
	id := info.getContainerID()
	if id == "" {
		logger.Debug("HOSTNAME environment variable is not set")
		return nil
	}
	inspect, err := info.ContainerInspecter.ContainerInspect(ctx, id)
	if err != nil {
		logger.Error("failed to inspect container", "error", err)
		return nil
	}
	info.container = &inspect
	return &inspect
}

func (info *DogowaiterDockerInfo) GetStackName(ctx context.Context) string {
	if info.stackName != "" {
		return info.stackName
	}
	c := info.getCurrentContainer(ctx)
	if c == nil {
		return ""
	}
	if c.Config != nil {
		if label, ok := c.Config.Labels[dockerComposeProjectLabel]; ok {
			info.stackName = label
		}
	}
	return info.stackName
}
