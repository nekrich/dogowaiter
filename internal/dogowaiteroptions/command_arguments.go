package dogowaiteroptions

// DogowaiterCommandArguments is the CLI (flags + env). Env vars override defaults; flags override env.
type DogowaiterCommandArguments struct {
	ConfigFile   string `long:"config" description:"Config file path (YAML). Empty = fallback to DOGOWAITER_DEPENDENCIES for dependencies. [$DOGOWAITER_CONFIG_FILE] (default: /config/dogowaiter.yaml)"`
	Dependencies string `long:"dependencies" description:"Comma-separated stack:service or service (used when no config file is provided). [$DOGOWAITER_DEPENDENCIES]"`
	HealthFile   string `long:"health-file" description:"Path for the JSON health file. [$DOGOWAITER_HEALTH_FILE] (default: /tmp/healthy)"`
	DockerHost   string `long:"docker-host" description:"Docker API endpoint. [$DOGOWAITER_DOCKER_HOST, $DOCKER_HOST] (default: unix:///var/run/docker.sock)"`
	LogLevel     string `long:"log-level" description:"Log level: debug, info, error. [$DOGOWAITER_LOG_LEVEL, $LOG_LEVEL] (default: info)"`
}

// DogowaiterCommandOptionsInterface implementation (Get prefix to avoid field/method name conflict).
func (c DogowaiterCommandArguments) GetConfigFile() string   { return c.ConfigFile }
func (c DogowaiterCommandArguments) GetDependencies() string { return c.Dependencies }
func (c DogowaiterCommandArguments) GetHealthFile() string   { return c.HealthFile }
func (c DogowaiterCommandArguments) GetDockerHost() string   { return c.DockerHost }
func (c DogowaiterCommandArguments) GetLogLevel() string     { return c.LogLevel }
