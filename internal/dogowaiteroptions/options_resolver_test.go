package dogowaiteroptions

import (
	"testing"
)

func TestDogowaiterOptionsResolver_Resolve_argument_overrides_env_and_default(t *testing.T) {
	args := DogowaiterCommandArguments{
		ConfigFile: "/arg-config.yaml",
		LogLevel:   "debug",
	}
	env := DogowaiterCommandEnv{
		ConfigFile: "/env-config.yaml",
		HealthFile: "/env-health",
	}
	defaults := NewDogowaiterCommandDefaults()
	configFile := &DogowaiterConfigurationFile{}

	resolver := DogowaiterOptionsResolver{}
	got := resolver.Resolve(args, env, configFile, defaults)

	// Argument wins
	if got.ConfigFile.Value != "/arg-config.yaml" || got.ConfigFile.Source != DogowaiterOptionSourceArgument {
		t.Errorf("ConfigFile: value=%q source=%q", got.ConfigFile.Value, got.ConfigFile.Source)
	}
	if got.LogLevel.Value != "debug" || got.LogLevel.Source != DogowaiterOptionSourceArgument {
		t.Errorf("LogLevel: value=%q source=%q", got.LogLevel.Value, got.LogLevel.Source)
	}
	// Env wins where args empty
	if got.HealthFile.Value != "/env-health" || got.HealthFile.Source != DogowaiterOptionSourceEnvironment {
		t.Errorf("HealthFile: value=%q source=%q", got.HealthFile.Value, got.HealthFile.Source)
	}
	// Default where both empty
	if got.DockerHost.Value != "unix:///var/run/docker.sock" || got.DockerHost.Source != DogowaiterOptionSourceDefault {
		t.Errorf("DockerHost: value=%q source=%q", got.DockerHost.Value, got.DockerHost.Source)
	}
}

func TestDogowaiterOptionsResolver_Resolve_env_overrides_default(t *testing.T) {
	args := DogowaiterCommandArguments{}
	env := DogowaiterCommandEnv{
		DockerHost: "tcp://env-docker:2375",
	}
	defaults := NewDogowaiterCommandDefaults()
	configFile := &DogowaiterConfigurationFile{}

	resolver := DogowaiterOptionsResolver{}
	got := resolver.Resolve(args, env, configFile, defaults)

	if got.DockerHost.Value != "tcp://env-docker:2375" || got.DockerHost.Source != DogowaiterOptionSourceEnvironment {
		t.Errorf("DockerHost: value=%q source=%q", got.DockerHost.Value, got.DockerHost.Source)
	}
	if got.HealthFile.Value != "/tmp/healthy" || got.HealthFile.Source != DogowaiterOptionSourceDefault {
		t.Errorf("HealthFile: value=%q source=%q", got.HealthFile.Value, got.HealthFile.Source)
	}
}

func TestDogowaiterOptionsResolver_Resolve_all_defaults(t *testing.T) {
	args := DogowaiterCommandArguments{}
	env := DogowaiterCommandEnv{}
	defaults := NewDogowaiterCommandDefaults()
	configFile := &DogowaiterConfigurationFile{}

	resolver := DogowaiterOptionsResolver{}
	got := resolver.Resolve(args, env, configFile, defaults)

	if got.ConfigFile.Value != defaults.GetConfigFile() || got.ConfigFile.Source != DogowaiterOptionSourceDefault {
		t.Errorf("ConfigFile: value=%q source=%q", got.ConfigFile.Value, got.ConfigFile.Source)
	}
	if len(got.Dependencies.Value) != 0 {
		t.Errorf("Dependencies: value=%v (want empty)", got.Dependencies.Value)
	}
	if got.HealthFile.Value != defaults.GetHealthFile() {
		t.Errorf("HealthFile: value=%q", got.HealthFile.Value)
	}
	if got.DockerHost.Value != defaults.GetDockerHost() {
		t.Errorf("DockerHost: value=%q", got.DockerHost.Value)
	}
	if got.LogLevel.Value != defaults.GetLogLevel() {
		t.Errorf("LogLevel: value=%q", got.LogLevel.Value)
	}
}

func TestDogowaiterOptionsResolver_Resolve_configFile_nil(t *testing.T) {
	args := DogowaiterCommandArguments{}
	env := DogowaiterCommandEnv{}
	defaults := NewDogowaiterCommandDefaults()

	resolver := DogowaiterOptionsResolver{}
	got := resolver.Resolve(args, env, nil, defaults)

	// Same as all-defaults: no panic, use defaults when no args/env
	if got.ConfigFile.Value != defaults.GetConfigFile() || got.ConfigFile.Source != DogowaiterOptionSourceDefault {
		t.Errorf("ConfigFile: value=%q source=%q", got.ConfigFile.Value, got.ConfigFile.Source)
	}
	if len(got.Dependencies.Value) != 0 {
		t.Errorf("Dependencies: value=%v (want empty)", got.Dependencies.Value)
	}
	if got.HealthFile.Value != defaults.GetHealthFile() {
		t.Errorf("HealthFile: value=%q", got.HealthFile.Value)
	}
	if got.DockerHost.Value != defaults.GetDockerHost() {
		t.Errorf("DockerHost: value=%q", got.DockerHost.Value)
	}
	if got.LogLevel.Value != defaults.GetLogLevel() {
		t.Errorf("LogLevel: value=%q", got.LogLevel.Value)
	}
}

func TestDogowaiterOptionsResolver_Resolve_configFile_nil_args_still_win(t *testing.T) {
	args := DogowaiterCommandArguments{
		DockerHost: "tcp://arg-docker:2375",
		LogLevel:   "error",
	}
	env := DogowaiterCommandEnv{}
	defaults := NewDogowaiterCommandDefaults()

	resolver := DogowaiterOptionsResolver{}
	got := resolver.Resolve(args, env, nil, defaults)

	if got.DockerHost.Value != "tcp://arg-docker:2375" || got.DockerHost.Source != DogowaiterOptionSourceArgument {
		t.Errorf("DockerHost: value=%q source=%q", got.DockerHost.Value, got.DockerHost.Source)
	}
	if got.LogLevel.Value != "error" || got.LogLevel.Source != DogowaiterOptionSourceArgument {
		t.Errorf("LogLevel: value=%q source=%q", got.LogLevel.Value, got.LogLevel.Source)
	}
}
