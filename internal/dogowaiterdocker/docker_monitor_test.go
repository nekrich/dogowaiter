package dogowaiterdocker

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/docker/docker/api/types/events"
)

func TestMustReloadConfig(t *testing.T) {
	tests := []struct {
		action events.Action
		want   bool
	}{
		{events.ActionCreate, true},
		{events.ActionDelete, true},
		{events.ActionStart, false},
		{events.ActionDie, false},
		{events.ActionHealthStatus, false},
		{events.Action("unknown"), false},
	}
	for _, tt := range tests {
		t.Run(string(tt.action), func(t *testing.T) {
			if got := mustReloadConfig(tt.action); got != tt.want {
				t.Errorf("mustReloadConfig(%q) = %v, want %v", tt.action, got, tt.want)
			}
		})
	}
}

func TestMustUpdateHealthCheckResult(t *testing.T) {
	tests := []struct {
		action events.Action
		want   bool
	}{
		{events.ActionStart, true},
		{events.ActionDie, true},
		{events.ActionHealthStatus, true},
		{events.ActionCreate, false},
		{events.ActionDelete, false},
		{events.Action("unknown"), false},
	}
	for _, tt := range tests {
		t.Run(string(tt.action), func(t *testing.T) {
			if got := mustUpdateHealthCheckResult(tt.action); got != tt.want {
				t.Errorf("mustUpdateHealthCheckResult(%q) = %v, want %v", tt.action, got, tt.want)
			}
		})
	}
}

func TestIsRelevantAction(t *testing.T) {
	relevant := []string{"create", "delete", "start", "die", "health_status"}
	for _, a := range relevant {
		if !isRelevantAction(a) {
			t.Errorf("isRelevantAction(%q) = false, want true", a)
		}
	}
	if isRelevantAction("attach") {
		t.Error("isRelevantAction(\"attach\") = true, want false")
	}
	if isRelevantAction("") {
		t.Error("isRelevantAction(\"\") = true, want false")
	}
}

func TestNewEventFilters(t *testing.T) {
	args := newEventFilters()
	if args.Len() == 0 {
		t.Fatal("expected non-empty filters")
	}
	// type=container
	typeVals := args.Get("type")
	if len(typeVals) == 0 {
		t.Fatal("expected type filter")
	}
	if typeVals[0] != string(events.ContainerEventType) {
		t.Errorf("filter type = %q, want %q", typeVals[0], events.ContainerEventType)
	}
	// event filter must include our actions
	eventVals := args.Get("event")
	if len(eventVals) == 0 {
		t.Fatal("expected event filter")
	}
	expectedEvents := []string{"create", "delete", "start", "die", "health_status"}
	for _, e := range expectedEvents {
		if !sliceContains(eventVals, e) {
			t.Errorf("filters event missing %q, got %v", e, eventVals)
		}
	}
}

func sliceContains(s []string, v string) bool {
	for _, x := range s {
		if x == v {
			return true
		}
	}
	return false
}

func TestReloadConfigActions(t *testing.T) {
	actions := reloadConfigActions()
	if len(actions) != 2 {
		t.Fatalf("reloadConfigActions() len = %d, want 2", len(actions))
	}
	if !sliceContains([]string{string(actions[0]), string(actions[1])}, "create") {
		t.Error("reloadConfigActions missing create")
	}
	if !sliceContains([]string{string(actions[0]), string(actions[1])}, "delete") {
		t.Error("reloadConfigActions missing delete")
	}
}

func TestUpdateHealthCheckResultActions(t *testing.T) {
	actions := updateHealthCheckResultActions()
	if len(actions) != 3 {
		t.Fatalf("updateHealthCheckResultActions() len = %d, want 3", len(actions))
	}
}

func TestRelevantActions(t *testing.T) {
	actions := relevantActions()
	// reload (2) + health (3) = 5, no duplicates
	if len(actions) != 5 {
		t.Errorf("relevantActions() len = %d, want 5", len(actions))
	}
}

// fakeDockerEventsSource implements DockerEventsSource for tests. First Events() call sends
// the configured messages then closes the event channel; later calls return channels that block forever.
type fakeDockerEventsSource struct {
	events    []events.Message
	callCount int
}

func (f *fakeDockerEventsSource) Events(ctx context.Context, opts events.ListOptions) (<-chan events.Message, <-chan error) {
	evCh := make(chan events.Message, len(f.events)+1)
	errCh := make(chan error, 1)
	f.callCount++
	if f.callCount == 1 {
		go func() {
			for _, e := range f.events {
				evCh <- e
			}
			close(evCh)
		}()
		return evCh, errCh
	}
	// block forever so monitor doesn't busy-loop after first batch
	return make(chan events.Message), make(chan error)
}

func (f *fakeDockerEventsSource) Close() error { return nil }

func containerEvent(action events.Action, id string) events.Message {
	return events.Message{
		Type:   events.ContainerEventType,
		Action: action,
		Actor:  events.Actor{ID: id},
	}
}

func TestMonitorDependencies_trigger_create(t *testing.T) {
	fake := &fakeDockerEventsSource{
		events: []events.Message{containerEvent(events.ActionCreate, "cid-create")},
	}
	monitor := &DogowaiterDockerMonitor{
		Configuration: DogowaiterDockerMonitorConfiguration{Logger: slog.Default()},
		EventSource:   fake,
	}
	triggerCh := monitor.MonitorDependencies()
	defer monitor.Close()

	select {
	case ev := <-triggerCh:
		if !ev.ReloadConfig {
			t.Error("ReloadConfig want true for create")
		}
		if ev.UpdateHealthCheckResult {
			t.Error("UpdateHealthCheckResult want false for create")
		}
		if ev.ContainerID != "cid-create" {
			t.Errorf("ContainerID = %q, want cid-create", ev.ContainerID)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("expected trigger for create event")
	}
}

func TestMonitorDependencies_trigger_delete(t *testing.T) {
	fake := &fakeDockerEventsSource{
		events: []events.Message{containerEvent(events.ActionDelete, "cid-delete")},
	}
	monitor := &DogowaiterDockerMonitor{
		Configuration: DogowaiterDockerMonitorConfiguration{Logger: slog.Default()},
		EventSource:   fake,
	}
	triggerCh := monitor.MonitorDependencies()
	defer monitor.Close()

	select {
	case ev := <-triggerCh:
		if !ev.ReloadConfig {
			t.Error("ReloadConfig want true for delete")
		}
		if ev.UpdateHealthCheckResult {
			t.Error("UpdateHealthCheckResult want false for delete")
		}
		if ev.ContainerID != "cid-delete" {
			t.Errorf("ContainerID = %q", ev.ContainerID)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("expected trigger for delete event")
	}
}

func TestMonitorDependencies_trigger_start(t *testing.T) {
	fake := &fakeDockerEventsSource{
		events: []events.Message{containerEvent(events.ActionStart, "cid-start")},
	}
	monitor := &DogowaiterDockerMonitor{
		Configuration: DogowaiterDockerMonitorConfiguration{Logger: slog.Default()},
		EventSource:   fake,
	}
	triggerCh := monitor.MonitorDependencies()
	defer monitor.Close()

	select {
	case ev := <-triggerCh:
		if ev.ReloadConfig {
			t.Error("ReloadConfig want false for start")
		}
		if !ev.UpdateHealthCheckResult {
			t.Error("UpdateHealthCheckResult want true for start")
		}
		if ev.ContainerID != "cid-start" {
			t.Errorf("ContainerID = %q", ev.ContainerID)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("expected trigger for start event")
	}
}

func TestMonitorDependencies_trigger_die(t *testing.T) {
	fake := &fakeDockerEventsSource{
		events: []events.Message{containerEvent(events.ActionDie, "cid-die")},
	}
	monitor := &DogowaiterDockerMonitor{
		Configuration: DogowaiterDockerMonitorConfiguration{Logger: slog.Default()},
		EventSource:   fake,
	}
	triggerCh := monitor.MonitorDependencies()
	defer monitor.Close()

	select {
	case ev := <-triggerCh:
		if ev.ReloadConfig {
			t.Error("ReloadConfig want false for die")
		}
		if !ev.UpdateHealthCheckResult {
			t.Error("UpdateHealthCheckResult want true for die")
		}
		if ev.ContainerID != "cid-die" {
			t.Errorf("ContainerID = %q", ev.ContainerID)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("expected trigger for die event")
	}
}

func TestMonitorDependencies_trigger_health_status(t *testing.T) {
	fake := &fakeDockerEventsSource{
		events: []events.Message{containerEvent(events.ActionHealthStatus, "cid-health")},
	}
	monitor := &DogowaiterDockerMonitor{
		Configuration: DogowaiterDockerMonitorConfiguration{Logger: slog.Default()},
		EventSource:   fake,
	}
	triggerCh := monitor.MonitorDependencies()
	defer monitor.Close()

	select {
	case ev := <-triggerCh:
		if ev.ReloadConfig {
			t.Error("ReloadConfig want false for health_status")
		}
		if !ev.UpdateHealthCheckResult {
			t.Error("UpdateHealthCheckResult want true for health_status")
		}
		if ev.ContainerID != "cid-health" {
			t.Errorf("ContainerID = %q", ev.ContainerID)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("expected trigger for health_status event")
	}
}

func TestMonitorDependencies_irrelevant_action_no_trigger(t *testing.T) {
	fake := &fakeDockerEventsSource{
		events: []events.Message{{
			Type:   events.ContainerEventType,
			Action: events.Action("attach"),
			Actor:  events.Actor{ID: "cid-attach"},
		}},
	}
	monitor := &DogowaiterDockerMonitor{
		Configuration: DogowaiterDockerMonitorConfiguration{Logger: slog.Default()},
		EventSource:   fake,
	}
	triggerCh := monitor.MonitorDependencies()
	defer monitor.Close()

	select {
	case <-triggerCh:
		t.Error("expected no trigger for attach event")
	case <-time.After(200 * time.Millisecond):
		// no trigger is correct
	}
}
