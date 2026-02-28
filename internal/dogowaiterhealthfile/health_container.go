package dogowaiterhealthfile

import "strings"

type HealthContainer struct {
	ContainerID string `json:"container_id"`
	Container   string `json:"container"`
	Reason      string `json:"reason"`
	IsReady     bool   `json:"is_ready"`
}

func (lhs HealthContainer) Equal(rhs HealthContainer) bool {
	if strings.Compare(lhs.ContainerID, rhs.ContainerID) != 0 {
		return false
	}
	if strings.Compare(lhs.Container, rhs.Container) != 0 {
		return false
	}
	if strings.Compare(lhs.Reason, rhs.Reason) != 0 {
		return false
	}
	if lhs.IsReady != rhs.IsReady {
		return false
	}
	return true
}

func (lhs HealthContainer) Compare(rhs HealthContainer) int {
	if lhs == rhs {
		return 0
	}
	// ready must go first
	if lhs.IsReady != rhs.IsReady {
		if lhs.IsReady {
			return -1
		}
		return 1
	}
	// then reason
	if lhs.Reason != rhs.Reason {
		return strings.Compare(lhs.Reason, rhs.Reason)
	}
	// then container
	if lhs.Container != rhs.Container {
		return strings.Compare(lhs.Container, rhs.Container)
	}
	// then container_id
	if lhs.ContainerID != rhs.ContainerID {
		return strings.Compare(lhs.ContainerID, rhs.ContainerID)
	}
	return 0
}
