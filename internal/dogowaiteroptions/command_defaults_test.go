package dogowaiteroptions

import "testing"

func TestNewDogowaiterCommandDefaults(t *testing.T) {
	got := NewDogowaiterCommandDefaults()
	want := DogowaiterCommandDefaults{
		ConfigFile: "config/dogowaiter.yaml",
		HealthFile: "/tmp/healthy",
		DockerHost: "unix:///var/run/docker.sock",
		LogLevel:   "info",
	}
	if got != want {
		t.Errorf("NewDogowaiterCommandDefaults() = %+v, want %+v", got, want)
	}
}

func TestDogowaiterCommandDefaults_getters(t *testing.T) {
	d := NewDogowaiterCommandDefaults()
	if d.GetConfigFile() != "config/dogowaiter.yaml" {
		t.Errorf("GetConfigFile() = %q, want config/dogowaiter.yaml", d.GetConfigFile())
	}
	if d.GetDependencies() != "" {
		t.Errorf("GetDependencies() = %q, want empty", d.GetDependencies())
	}
	if d.GetHealthFile() != "/tmp/healthy" {
		t.Errorf("GetHealthFile() = %q, want /tmp/healthy", d.GetHealthFile())
	}
	if d.GetDockerHost() != "unix:///var/run/docker.sock" {
		t.Errorf("GetDockerHost() = %q, want unix:///var/run/docker.sock", d.GetDockerHost())
	}
	if d.GetLogLevel() != "info" {
		t.Errorf("GetLogLevel() = %q, want info", d.GetLogLevel())
	}
}
