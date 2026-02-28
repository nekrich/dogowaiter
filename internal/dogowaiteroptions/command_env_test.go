package dogowaiteroptions

import (
	"testing"
)

func TestNewDogowaiterCommandEnv_loadsDOGOWAITER_vars(t *testing.T) {
	t.Setenv("DOGOWAITER_CONFIG_FILE", "/my/config.yaml")
	t.Setenv("DOGOWAITER_DEPENDENCIES", "stack:svc")
	t.Setenv("DOGOWAITER_HEALTH_FILE", "/tmp/health")
	t.Setenv("DOGOWAITER_DOCKER_HOST", "tcp://host:2375")
	t.Setenv("DOGOWAITER_LOG_LEVEL", "debug")

	env := NewDogowaiterCommandEnv()

	if env.GetConfigFile() != "/my/config.yaml" {
		t.Errorf("GetConfigFile() = %q, want /my/config.yaml", env.GetConfigFile())
	}
	if env.GetDependencies() != "stack:svc" {
		t.Errorf("GetDependencies() = %q, want stack:svc", env.GetDependencies())
	}
	if env.GetHealthFile() != "/tmp/health" {
		t.Errorf("GetHealthFile() = %q, want /tmp/health", env.GetHealthFile())
	}
	if env.GetDockerHost() != "tcp://host:2375" {
		t.Errorf("GetDockerHost() = %q, want tcp://host:2375", env.GetDockerHost())
	}
	if env.GetLogLevel() != "debug" {
		t.Errorf("GetLogLevel() = %q, want debug", env.GetLogLevel())
	}
}

func TestNewDogowaiterCommandEnv_resolveEnvDefaults_fallback(t *testing.T) {
	// No DOGOWAITER_* set; fall back to LOG_LEVEL and DOCKER_HOST
	t.Setenv("LOG_LEVEL", "warn")
	t.Setenv("DOCKER_HOST", "unix:///other/docker.sock")

	env := NewDogowaiterCommandEnv()

	if env.GetLogLevel() != "warn" {
		t.Errorf("GetLogLevel() = %q, want warn (from LOG_LEVEL)", env.GetLogLevel())
	}
	if env.GetDockerHost() != "unix:///other/docker.sock" {
		t.Errorf("GetDockerHost() = %q, want unix:///other/docker.sock (from DOCKER_HOST)", env.GetDockerHost())
	}
}

func TestNewDogowaiterCommandEnv_DOGOWAITER_overrides_fallback(t *testing.T) {
	t.Setenv("DOGOWAITER_LOG_LEVEL", "error")
	t.Setenv("LOG_LEVEL", "debug")
	t.Setenv("DOGOWAITER_DOCKER_HOST", "tcp://docker:2375")
	t.Setenv("DOCKER_HOST", "unix:///fallback.sock")

	env := NewDogowaiterCommandEnv()

	if env.GetLogLevel() != "error" {
		t.Errorf("GetLogLevel() = %q, want error (DOGOWAITER_* takes precedence)", env.GetLogLevel())
	}
	if env.GetDockerHost() != "tcp://docker:2375" {
		t.Errorf("GetDockerHost() = %q, want tcp://docker:2375", env.GetDockerHost())
	}
}
