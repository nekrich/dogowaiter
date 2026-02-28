package dogowaiterdockercurrent

import (
	"context"
	"errors"
	"log/slog"
	"testing"

	"github.com/docker/docker/api/types/container"
)

// fakeContainerInspecter implements ContainerInspecter for tests.
type fakeContainerInspector struct {
	response container.InspectResponse
	err      error
}

func (f *fakeContainerInspector) ContainerInspect(ctx context.Context, containerID string) (container.InspectResponse, error) {
	if f.err != nil {
		return container.InspectResponse{}, f.err
	}
	return f.response, nil
}

// fakeContainerIDProvider implements ContainerIDProvider for tests.
type fakeContainerIDProvider struct {
	id string
}

func (f *fakeContainerIDProvider) ContainerID() string {
	return f.id
}

func TestGetStackName_returnsCachedStackName(t *testing.T) {
	info := &DogowaiterDockerInfo{
		ContainerIDProvider: &fakeContainerIDProvider{},
		Logger:              slog.Default(),
	}
	info.stackName = "cached_stack"

	ctx := context.Background()
	got := info.GetStackName(ctx)
	if got != "cached_stack" {
		t.Errorf("GetStackName() = %q, want %q", got, "cached_stack")
	}
}

func TestGetStackName_returnsEmptyWhenContainerIDEmpty(t *testing.T) {
	info := &DogowaiterDockerInfo{
		ContainerInspecter:  &fakeContainerInspector{},
		ContainerIDProvider: &fakeContainerIDProvider{id: ""},
		Logger:              slog.Default(),
	}

	ctx := context.Background()
	got := info.GetStackName(ctx)
	if got != "" {
		t.Errorf("GetStackName() = %q, want \"\"", got)
	}
}

func TestGetStackName_returnsEmptyWhenInspectFails(t *testing.T) {
	info := &DogowaiterDockerInfo{
		ContainerInspecter:  &fakeContainerInspector{err: errors.New("inspect failed")},
		ContainerIDProvider: &fakeContainerIDProvider{id: "some-id"},
		Logger:              slog.Default(),
	}

	ctx := context.Background()
	got := info.GetStackName(ctx)
	if got != "" {
		t.Errorf("GetStackName() = %q, want \"\"", got)
	}
}

func TestGetStackName_fromContainerLabels(t *testing.T) {
	info := &DogowaiterDockerInfo{
		ContainerInspecter: &fakeContainerInspector{
			response: container.InspectResponse{
				Config: &container.Config{
					Labels: map[string]string{dockerComposeProjectLabel: "myproject"},
				},
			},
		},
		ContainerIDProvider: &fakeContainerIDProvider{id: "test-id"},
		Logger:              slog.Default(),
	}

	ctx := context.Background()
	got := info.GetStackName(ctx)
	if got != "myproject" {
		t.Errorf("GetStackName() = %q, want %q", got, "myproject")
	}
	// second call returns cached (no second inspect)
	if info.GetStackName(ctx) != "myproject" {
		t.Error("second GetStackName() should return cached value")
	}
}

func TestGetStackName_returnsEmptyWhenLabelMissing(t *testing.T) {
	info := &DogowaiterDockerInfo{
		ContainerInspecter: &fakeContainerInspector{
			response: container.InspectResponse{
				Config: &container.Config{
					Labels: map[string]string{"other": "label"},
				},
			},
		},
		ContainerIDProvider: &fakeContainerIDProvider{id: "test-id"},
		Logger:              slog.Default(),
	}

	ctx := context.Background()
	got := info.GetStackName(ctx)
	if got != "" {
		t.Errorf("GetStackName() = %q, want \"\"", got)
	}
}

func TestGetStackName_returnsEmptyWhenConfigNil(t *testing.T) {
	info := &DogowaiterDockerInfo{
		ContainerInspecter: &fakeContainerInspector{
			response: container.InspectResponse{Config: nil},
		},
		ContainerIDProvider: &fakeContainerIDProvider{id: "test-id"},
		Logger:              slog.Default(),
	}

	ctx := context.Background()
	got := info.GetStackName(ctx)
	if got != "" {
		t.Errorf("GetStackName() with nil Config = %q, want \"\"", got)
	}
}

func TestGetStackName_usesContainerIDProvider(t *testing.T) {
	info := &DogowaiterDockerInfo{
		ContainerInspecter: &fakeContainerInspector{
			response: container.InspectResponse{
				Config: &container.Config{
					Labels: map[string]string{dockerComposeProjectLabel: "from-provider"},
				},
			},
		},
		ContainerIDProvider: &fakeContainerIDProvider{id: "provider-id"},
		Logger:              slog.Default(),
	}

	ctx := context.Background()
	got := info.GetStackName(ctx)
	if got != "from-provider" {
		t.Errorf("GetStackName() = %q, want %q", got, "from-provider")
	}
}

func TestGetCurrentContainer_returnsCached(t *testing.T) {
	cached := &container.InspectResponse{
		Config: &container.Config{Labels: map[string]string{"x": "y"}},
	}
	info := &DogowaiterDockerInfo{
		ContainerIDProvider: &fakeContainerIDProvider{},
		Logger:              slog.Default(),
		container:           cached,
	}

	ctx := context.Background()
	got := info.getCurrentContainer(ctx)
	if got != cached {
		t.Error("getCurrentContainer() should return cached container")
	}
	if got.Config == nil || got.Config.Labels["x"] != "y" {
		t.Error("getCurrentContainer() should return cached container with same config")
	}
}

func TestGetCurrentContainer_returnsNilWhenNoInspecter(t *testing.T) {
	info := &DogowaiterDockerInfo{
		ContainerInspecter:  nil,
		ContainerIDProvider: &fakeContainerIDProvider{id: "id"},
		Logger:              slog.Default(),
	}

	ctx := context.Background()
	got := info.getCurrentContainer(ctx)
	if got != nil {
		t.Error("getCurrentContainer() with nil ContainerInspecter should return nil")
	}
}
