package dogowaiterhealthfile

import "slices"

// HealthCheckResult is the in-memory state: full list of monitored containers and overall healthy.
type HealthCheckResult struct {
	healthy    bool              `json:"-"`
	containers []HealthContainer `json:"-"`
}

func BuildHealthCheckResult(healthy bool, containers []HealthContainer) HealthCheckResult {
	clonedContainers := slices.Clone(containers)
	slices.SortFunc(clonedContainers, func(lhs, rhs HealthContainer) int { return lhs.Compare(rhs) })
	return HealthCheckResult{healthy: healthy, containers: clonedContainers}
}

func (h *HealthCheckResult) IsHealthy() bool {
	return h.healthy
}

func (h *HealthCheckResult) GetContainers() []HealthContainer {
	return h.containers
}

// Equal reports whether r and other represent the same health state.
func (lhs *HealthCheckResult) Equal(rhs *HealthCheckResult) bool {
	if rhs == nil || lhs == nil {
		return rhs == lhs
	}
	if lhs.healthy != rhs.healthy {
		return false
	}
	if len(lhs.containers) != len(rhs.containers) {
		return false
	}
	for i := range lhs.containers {
		lhsContainer, rhsContainer := lhs.containers[i], rhs.containers[i]
		if !lhsContainer.Equal(rhsContainer) {
			return false
		}
	}
	return true
}
