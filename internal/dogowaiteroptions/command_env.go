package dogowaiteroptions

import (
	"log/slog"
	"os"

	"github.com/caarlos0/env/v11"
)

// DogowaiterEnv holds options resolved only from environment variables (no flags/args).
// Use it to distinguish values set via CLI vs env-only. LoadFromOS parses via env tags.
type DogowaiterCommandEnv struct {
	ConfigFile   string `env:"DOGOWAITER_CONFIG_FILE" envDefault:""`
	Dependencies string `env:"DOGOWAITER_DEPENDENCIES" envDefault:""`
	HealthFile   string `env:"DOGOWAITER_HEALTH_FILE" envDefault:""`
	DockerHost   string `env:"DOGOWAITER_DOCKER_HOST" envDefault:""`
	LogLevel     string `env:"DOGOWAITER_LOG_LEVEL" envDefault:""`
}

// DogowaiterCommandOptionsInterface implementation (Get prefix to avoid field/method name conflict).
func (e DogowaiterCommandEnv) GetConfigFile() string   { return e.ConfigFile }
func (e DogowaiterCommandEnv) GetDependencies() string { return e.Dependencies }
func (e DogowaiterCommandEnv) GetHealthFile() string   { return e.HealthFile }
func (e DogowaiterCommandEnv) GetDockerHost() string   { return e.DockerHost }
func (e DogowaiterCommandEnv) GetLogLevel() string     { return e.LogLevel }

// loadFromEnv fills the struct from environment variables using the env struct tags.
func (commandEnv *DogowaiterCommandEnv) loadFromEnv() error {
	return env.Parse(commandEnv)
}

// ResolveEnvDefaults fills empty fields with fallback env vars.
func (commandEnv *DogowaiterCommandEnv) resolveEnvDefaults() {
	if commandEnv.LogLevel == "" {
		commandEnv.LogLevel = os.Getenv("LOG_LEVEL")
	}
	if commandEnv.DockerHost == "" {
		commandEnv.DockerHost = os.Getenv("DOCKER_HOST")
	}
}

// NewDogowaiterCommandEnv loads the environment variables into a DogowaiterCommandEnv and resolves (fallbacks + defaults).
// Parse errors from env are ignored; the struct is still resolved.
func NewDogowaiterCommandEnv() DogowaiterCommandEnv {
	var commandEnv DogowaiterCommandEnv
	err := commandEnv.loadFromEnv()
	if err != nil {
		slog.Error("failed to load environment variables", "error", err)
		return commandEnv
	}
	commandEnv.resolveEnvDefaults()
	return commandEnv
}
