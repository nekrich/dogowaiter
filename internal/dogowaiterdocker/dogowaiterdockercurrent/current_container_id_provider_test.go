package dogowaiterdockercurrent

import (
	"testing"
)

func TestDogowaiterDockerCurrentContainerIDProvider_ContainerID(t *testing.T) {
	p := DogowaiterDockerCurrentContainerIDProvider{}

	t.Run("returns HOSTNAME when set", func(t *testing.T) {
		t.Setenv("HOSTNAME", "abc123")
		got := p.ContainerID()
		if got != "abc123" {
			t.Errorf("ContainerID() = %q, want %q", got, "abc123")
		}
	})

	t.Run("returns empty when HOSTNAME unset", func(t *testing.T) {
		t.Setenv("HOSTNAME", "")
		got := p.ContainerID()
		if got != "" {
			t.Errorf("ContainerID() = %q, want \"\"", got)
		}
	})
}
