package dogowaiterhealthfile

import (
	"testing"
)

func TestBuildHealthCheckResult(t *testing.T) {
	containers := []HealthContainer{
		{ContainerID: "z", Container: "z", Reason: "z", IsReady: false},
		{ContainerID: "a", Container: "a", Reason: "a", IsReady: true},
	}
	result := BuildHealthCheckResult(true, containers)

	if !result.IsHealthy() {
		t.Error("IsHealthy() want true")
	}
	got := result.GetContainers()
	if len(got) != 2 {
		t.Fatalf("GetContainers() len = %d, want 2", len(got))
	}
	// BuildHealthCheckResult sorts: ready first, then reason, container, container_id
	if !got[0].IsReady || got[0].Container != "a" {
		t.Errorf("first container = %+v, want ready a", got[0])
	}
	if got[1].IsReady || got[1].Container != "z" {
		t.Errorf("second container = %+v, want not ready z", got[1])
	}
	// clone: original slice unchanged order
	if containers[0].Container != "z" {
		t.Error("BuildHealthCheckResult should not mutate input slice order")
	}
}

func TestBuildHealthCheckResult_empty(t *testing.T) {
	result := BuildHealthCheckResult(false, nil)
	if result.IsHealthy() {
		t.Error("IsHealthy() want false")
	}
	if result.GetContainers() != nil {
		t.Errorf("GetContainers() = %v, want nil", result.GetContainers())
	}
}

func TestHealthCheckResult_Equal(t *testing.T) {
	a := BuildHealthCheckResult(true, []HealthContainer{{ContainerID: "1", Container: "c", Reason: "r", IsReady: true}})
	b := BuildHealthCheckResult(true, []HealthContainer{{ContainerID: "1", Container: "c", Reason: "r", IsReady: true}})
	c := BuildHealthCheckResult(false, []HealthContainer{{ContainerID: "1", Container: "c", Reason: "r", IsReady: true}})
	d := BuildHealthCheckResult(true, []HealthContainer{{ContainerID: "2", Container: "c", Reason: "r", IsReady: true}})

	if !a.Equal(&b) {
		t.Error("a.Equal(b) want true")
	}
	if a.Equal(&c) {
		t.Error("a.Equal(c) want false (different healthy)")
	}
	if a.Equal(&d) {
		t.Error("a.Equal(d) want false (different containers)")
	}
	if a.Equal(nil) {
		t.Error("a.Equal(nil) want false (receiver non-nil)")
	}
	var nilResult *HealthCheckResult
	if !nilResult.Equal(nil) {
		t.Error("nil.Equal(nil) want true")
	}
	if nilResult.Equal(&a) {
		t.Error("nil.Equal(a) want false")
	}
}
