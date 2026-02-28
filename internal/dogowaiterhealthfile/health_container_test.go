package dogowaiterhealthfile

import (
	"testing"
)

func TestHealthContainer_Equal(t *testing.T) {
	a := HealthContainer{ContainerID: "id1", Container: "c1", Reason: "r1", IsReady: true}
	b := HealthContainer{ContainerID: "id1", Container: "c1", Reason: "r1", IsReady: true}
	c := HealthContainer{ContainerID: "id2", Container: "c1", Reason: "r1", IsReady: true}
	d := HealthContainer{ContainerID: "id1", Container: "c2", Reason: "r1", IsReady: true}
	e := HealthContainer{ContainerID: "id1", Container: "c1", Reason: "r2", IsReady: true}
	f := HealthContainer{ContainerID: "id1", Container: "c1", Reason: "r1", IsReady: false}

	if !a.Equal(b) {
		t.Error("a.Equal(b) want true")
	}
	if a.Equal(c) {
		t.Error("a.Equal(c) want false (different ContainerID)")
	}
	if a.Equal(d) {
		t.Error("a.Equal(d) want false (different Container)")
	}
	if a.Equal(e) {
		t.Error("a.Equal(e) want false (different Reason)")
	}
	if a.Equal(f) {
		t.Error("a.Equal(f) want false (different IsReady)")
	}
}

func TestHealthContainer_Compare(t *testing.T) {
	ready := HealthContainer{ContainerID: "b", Container: "b", Reason: "b", IsReady: true}
	notReady := HealthContainer{ContainerID: "a", Container: "a", Reason: "a", IsReady: false}

	// ready must sort before not ready
	if got := ready.Compare(notReady); got >= 0 {
		t.Errorf("ready.Compare(notReady) = %d, want < 0", got)
	}
	if got := notReady.Compare(ready); got <= 0 {
		t.Errorf("notReady.Compare(ready) = %d, want > 0", got)
	}

	// same IsReady: compare by Reason
	r1 := HealthContainer{ContainerID: "x", Container: "x", Reason: "a", IsReady: true}
	r2 := HealthContainer{ContainerID: "x", Container: "x", Reason: "b", IsReady: true}
	if got := r1.Compare(r2); got >= 0 {
		t.Errorf("r1.Compare(r2) = %d, want < 0", got)
	}

	// same Reason: compare by Container
	c1 := HealthContainer{ContainerID: "x", Container: "a", Reason: "r", IsReady: true}
	c2 := HealthContainer{ContainerID: "x", Container: "b", Reason: "r", IsReady: true}
	if got := c1.Compare(c2); got >= 0 {
		t.Errorf("c1.Compare(c2) = %d, want < 0", got)
	}

	// same Container: compare by ContainerID
	id1 := HealthContainer{ContainerID: "a", Container: "c", Reason: "r", IsReady: true}
	id2 := HealthContainer{ContainerID: "b", Container: "c", Reason: "r", IsReady: true}
	if got := id1.Compare(id2); got >= 0 {
		t.Errorf("id1.Compare(id2) = %d, want < 0", got)
	}

	// equal containers
	if got := id1.Compare(id1); got != 0 {
		t.Errorf("id1.Compare(id1) = %d, want 0", got)
	}
}
