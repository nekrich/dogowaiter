package dogowaiterdocker

import (
	"testing"
)

func TestReasonConstants(t *testing.T) {
	if ReasonNotStarted != "not started" {
		t.Errorf("ReasonNotStarted = %q, want \"not started\"", ReasonNotStarted)
	}
	if ReasonUnhealthy != "unhealthy" {
		t.Errorf("ReasonUnhealthy = %q, want \"unhealthy\"", ReasonUnhealthy)
	}
	if ReasonNotFound != "not found" {
		t.Errorf("ReasonNotFound = %q, want \"not found\"", ReasonNotFound)
	}
}

func TestCheckResult_fields(t *testing.T) {
	// CheckResult is a data struct; we only test that zero value is usable
	r := CheckResult{}
	if r.Pass {
		t.Error("zero CheckResult.Pass should be false")
	}
	r.Pass = true
	r.ContainerName = "svc"
	r.Reason = ReasonUnhealthy
	if r.ContainerName != "svc" || r.Reason != ReasonUnhealthy {
		t.Errorf("CheckResult = %+v", r)
	}
}
