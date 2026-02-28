package dogowaiterdocker

import (
	"testing"

	"dogowaiter/internal/dogowaiteroptions"

	"github.com/docker/docker/api/types/container"
)

func containerWithLabels(names []string, labels map[string]string) container.Summary {
	if names == nil {
		names = []string{"/default"}
	}
	return container.Summary{Names: names, Labels: labels}
}

func TestMatchContainer_byServiceLabel(t *testing.T) {
	dep := dogowaiteroptions.DogowaiterConfigurationDependency{Name: "web", Stack: ""}
	c := containerWithLabels([]string{"/stack_web_1"}, map[string]string{
		dockerComposeServiceLabel: "web",
		dockerComposeProjectLabel: "stack",
	})
	if !matchContainer(dep, c) {
		t.Error("matchContainer(web, service=web) want true")
	}
}

func TestMatchContainer_byName(t *testing.T) {
	dep := dogowaiteroptions.DogowaiterConfigurationDependency{Name: "web", Stack: ""}
	c := containerWithLabels(nil, nil)
	c.Names = []string{"/web"}
	if !matchContainer(dep, c) {
		t.Error("matchContainer(web, name=/web) want true")
	}
}

func TestMatchContainer_byNameContains(t *testing.T) {
	dep := dogowaiteroptions.DogowaiterConfigurationDependency{Name: "web", Stack: ""}
	c := containerWithLabels([]string{"/stack_web_1"}, nil)
	if !matchContainer(dep, c) {
		t.Error("matchContainer(web, name contains web) want true")
	}
}

func TestMatchContainer_stackMismatch(t *testing.T) {
	dep := dogowaiteroptions.DogowaiterConfigurationDependency{Name: "web", Stack: "stackA"}
	c := containerWithLabels([]string{"/stackB_web_1"}, map[string]string{
		dockerComposeServiceLabel: "web",
		dockerComposeProjectLabel: "stackB",
	})
	if matchContainer(dep, c) {
		t.Error("matchContainer(stackA/web, project=stackB) want false")
	}
}

func TestMatchContainer_stackMatch(t *testing.T) {
	dep := dogowaiteroptions.DogowaiterConfigurationDependency{Name: "web", Stack: "mystack"}
	c := containerWithLabels([]string{"/x"}, map[string]string{
		dockerComposeServiceLabel: "web",
		dockerComposeProjectLabel: "mystack",
	})
	if !matchContainer(dep, c) {
		t.Error("matchContainer(mystack/web, project=mystack) want true")
	}
}

func TestMatchContainer_noMatch(t *testing.T) {
	dep := dogowaiteroptions.DogowaiterConfigurationDependency{Name: "api", Stack: ""}
	c := containerWithLabels([]string{"/other"}, map[string]string{
		dockerComposeServiceLabel: "web",
	})
	if matchContainer(dep, c) {
		t.Error("matchContainer(api, service=web) want false")
	}
}

func TestDockerContainerResolveResult_Equal(t *testing.T) {
	a := DockerContainerResolveResult{DependencyName: "web", ContainerIDs: []string{"id1"}, ContainerNames: []string{"web"}}
	b := DockerContainerResolveResult{DependencyName: "web", ContainerIDs: []string{"id1"}, ContainerNames: []string{"web"}}
	c := DockerContainerResolveResult{DependencyName: "web", ContainerIDs: []string{"id2"}, ContainerNames: []string{"web"}}
	d := DockerContainerResolveResult{DependencyName: "api", ContainerIDs: []string{"id1"}, ContainerNames: []string{"web"}}

	if !a.Equal(b) {
		t.Error("a.Equal(b) want true")
	}
	if a.Equal(c) {
		t.Error("a.Equal(c) want false (different IDs)")
	}
	if a.Equal(d) {
		t.Error("a.Equal(d) want false (different name)")
	}
}

func TestGetWatchedContainerIDs(t *testing.T) {
	results := []DockerContainerResolveResult{
		{DependencyName: "a", ContainerIDs: []string{"id1", "id2"}, ContainerNames: nil},
		{DependencyName: "b", ContainerIDs: []string{"id2", "id3"}, ContainerNames: nil},
	}
	got := GetWatchedContainerIDs(results)
	want := map[string]struct{}{"id1": {}, "id2": {}, "id3": {}}
	if len(got) != len(want) {
		t.Errorf("GetWatchedContainerIDs len = %d, want 3", len(got))
	}
	for id := range want {
		if _, ok := got[id]; !ok {
			t.Errorf("GetWatchedContainerIDs missing %q", id)
		}
	}
}

func TestGetWatchedContainerIDs_empty(t *testing.T) {
	got := GetWatchedContainerIDs(nil)
	if len(got) != 0 {
		t.Errorf("GetWatchedContainerIDs(nil) = %v", got)
	}
	got = GetWatchedContainerIDs([]DockerContainerResolveResult{})
	if len(got) != 0 {
		t.Errorf("GetWatchedContainerIDs(empty) = %v", got)
	}
}
