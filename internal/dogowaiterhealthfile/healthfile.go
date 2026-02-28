package dogowaiterhealthfile

import (
	"encoding/json"
	"os"
	"slices"
	"strings"
)

type DogowaiterHealthFile struct {
	FilePath          string
	healthCheckResult *HealthCheckResult
}

func (h *DogowaiterHealthFile) Write(healthCheckResult *HealthCheckResult) error {
	if h.healthCheckResult.Equal(healthCheckResult) {
		return nil
	}
	h.healthCheckResult = healthCheckResult
	return writeHealthFile(h.FilePath, h.healthCheckResult)
}

func (h *DogowaiterHealthFile) Remove() error {
	return removeHealthFile(h.FilePath)
}

// fileContainer is the JSON shape for one container in the health file (container, reason, is_ready).
type fileContainer struct {
	Container string `json:"container"`
	IsReady   bool   `json:"is_ready"`
	Reason    string `json:"reason,omitempty"`
}

func (lhs fileContainer) Compare(rhs fileContainer) int {
	if lhs == rhs {
		return 0
	}
	// not is_ready must go first
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
	return strings.Compare(lhs.Container, rhs.Container)
}

// WriteHealthFile writes the health status JSON to path. Outputs "healthy" and "containers" (all containers with container, reason, is_ready).
func writeHealthFile(path string, result *HealthCheckResult) error {
	payload := make(map[string]interface{})
	payload["healthy"] = result.IsHealthy()
	var list []fileContainer
	for _, container := range result.GetContainers() {
		list = append(list, fileContainer{Container: container.Container, IsReady: container.IsReady, Reason: container.Reason})
	}
	slices.SortFunc(list, func(lhs, rhs fileContainer) int { return lhs.Compare(rhs) })
	payload["containers"] = list
	data, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

// RemoveHealthFile removes the health file at path, e.g. for clean shutdown on SIGTERM.
func removeHealthFile(path string) error {
	return os.Remove(path)
}
