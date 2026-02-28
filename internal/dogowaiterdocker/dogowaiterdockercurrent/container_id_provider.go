// DogowaiterDockerContainerIDProvider returns the current container ID. Used to discover "this" container.
package dogowaiterdockercurrent

type DogowaiterDockerContainerIDProvider interface {
	ContainerID() string
}
