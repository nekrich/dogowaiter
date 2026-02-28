// DogowaiterDockerContainerIDProvider returns the current container ID. Used to discover "this" container.
package dogowaiterdockercurrent

import "os"

// DogowaiterDockerCurrentContainerIDProvider provides the container ID from the HOSTNAME environment variable.
type DogowaiterDockerCurrentContainerIDProvider struct{}

func (p DogowaiterDockerCurrentContainerIDProvider) ContainerID() string {
	return os.Getenv("HOSTNAME")
}
