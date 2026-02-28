// DogowaiterDockerContainerInspector inspects a container by ID. *client.Client implements this interface.
package dogowaiterdockercurrent

import (
	"context"

	"github.com/docker/docker/api/types/container"
)

// DogowaiterDockerContainerInspector inspects a container by ID. *client.Client implements this interface.
type DogowaiterDockerContainerInspector interface {
	ContainerInspect(ctx context.Context, containerID string) (container.InspectResponse, error)
}
