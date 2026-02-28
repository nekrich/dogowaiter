package dogowaiteroptions

// DogowaiterCommandDefaults holds the default values for the command.
type DogowaiterCommandDefaults struct {
	ConfigFile string `default:"config/dogowaiter.yaml"`
	HealthFile string `default:"/tmp/healthy"`
	DockerHost string `default:"unix:///var/run/docker.sock"`
	LogLevel   string `default:"info"`
}

// DogowaiterCommandOptionsInterface implementation (Get prefix to avoid field/method name conflict).
func (e DogowaiterCommandDefaults) GetConfigFile() string   { return e.ConfigFile }
func (e DogowaiterCommandDefaults) GetDependencies() string { return "" }
func (e DogowaiterCommandDefaults) GetHealthFile() string   { return e.HealthFile }
func (e DogowaiterCommandDefaults) GetDockerHost() string   { return e.DockerHost }
func (e DogowaiterCommandDefaults) GetLogLevel() string     { return e.LogLevel }

// NewDogowaiterCommandDefaults creates a new DogowaiterCommandDefaults instance.
func NewDogowaiterCommandDefaults() DogowaiterCommandDefaults {
	return DogowaiterCommandDefaults{
		ConfigFile: "config/dogowaiter.yaml",
		HealthFile: "/tmp/healthy",
		DockerHost: "unix:///var/run/docker.sock",
		LogLevel:   "info",
	}
}
