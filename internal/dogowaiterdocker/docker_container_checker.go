package dogowaiterdocker

import (
	"context"
	"log/slog"
	"strings"

	"github.com/docker/docker/client"
)

const (
	// ReasonNotStarted is the failure reason when the container is not running or not found.
	ReasonNotStarted = "not started"
	// ReasonUnhealthy is the failure reason when the container is running but not healthy.
	ReasonUnhealthy = "unhealthy"
	// ReasonNotFound is the failure reason when the container is not found.
	ReasonNotFound = "not found"
)

// CheckResult is the result of checking one container.
type CheckResult struct {
	ContainerName string
	Pass          bool
	Reason        string
}

type DogowaiterDockerContainerCheckerInterface interface {
	CheckContainer(ctx context.Context, containerID string) (CheckResult, error)
}

type DogowaiterDockerContainerChecker struct {
	DockerClient *client.Client
	Logger       *slog.Logger
}

// CheckContainer inspects the container and returns pass/fail and reason. Ready is determined automatically:
// running and (no health config or health status == "healthy").
func (checker DogowaiterDockerContainerChecker) CheckContainer(ctx context.Context, containerID string) (CheckResult, error) {
	inspect, err := checker.DockerClient.ContainerInspect(ctx, containerID)
	if err != nil {
		return CheckResult{ContainerName: containerID, Pass: false, Reason: ReasonNotFound}, err
	}

	name := strings.TrimPrefix(inspect.Name, "/")

	state := inspect.State
	if state == nil || !state.Running {
		return CheckResult{ContainerName: name, Pass: false, Reason: ReasonNotStarted}, nil
	}

	// No health config: running is enough
	if state.Health == nil {
		return CheckResult{ContainerName: name, Pass: true}, nil
	}
	if state.Health.Status == "healthy" {
		return CheckResult{ContainerName: name, Pass: true}, nil
	}
	return CheckResult{ContainerName: name, Pass: false, Reason: ReasonUnhealthy}, nil
}
