package dogowaiteroptions

import "testing"

func TestDogowaiterCommandArguments_getters(t *testing.T) {
	c := DogowaiterCommandArguments{
		ConfigFile:   "/etc/dogowaiter.yaml",
		Dependencies: "stack:svc",
		HealthFile:   "/tmp/health.json",
		DockerHost:   "tcp://localhost:2375",
		LogLevel:     "debug",
	}
	if c.GetConfigFile() != "/etc/dogowaiter.yaml" {
		t.Errorf("GetConfigFile() = %q", c.GetConfigFile())
	}
	if c.GetDependencies() != "stack:svc" {
		t.Errorf("GetDependencies() = %q", c.GetDependencies())
	}
	if c.GetHealthFile() != "/tmp/health.json" {
		t.Errorf("GetHealthFile() = %q", c.GetHealthFile())
	}
	if c.GetDockerHost() != "tcp://localhost:2375" {
		t.Errorf("GetDockerHost() = %q", c.GetDockerHost())
	}
	if c.GetLogLevel() != "debug" {
		t.Errorf("GetLogLevel() = %q", c.GetLogLevel())
	}
}
