package dogowaiteroptions

import "reflect"

// DogowaiterOptionSource is the source of the option.
// One of the following sources is expected: arguments, environment, config_file.
type DogowaiterOptionSource string

const (
	DogowaiterOptionSourceDefault     DogowaiterOptionSource = "default"     // Pre-configured default value
	DogowaiterOptionSourceEnvironment DogowaiterOptionSource = "environment" // Environment variable
	DogowaiterOptionSourceConfigFile  DogowaiterOptionSource = "config_file" // Config file
	DogowaiterOptionSourceArgument    DogowaiterOptionSource = "argument"    // CLI argument
)

// DogowaiterOption is the option itself with its source and value.
type DogowaiterOption[T any] struct {
	Source DogowaiterOptionSource
	Value  T
}

// DogowaiterOptions are the resolved options for the dogowaiter application.
type DogowaiterOptions struct {
	ConfigFile   DogowaiterOption[string]                              // config file path
	Dependencies DogowaiterOption[[]DogowaiterConfigurationDependency] // dependencies
	HealthFile   DogowaiterOption[string]                              // health file path
	DockerHost   DogowaiterOption[string]                              // Docker API endpoint
	LogLevel     DogowaiterOption[string]                              // log level
}

// IsSet returns true if the option is set.
func (option DogowaiterOption[T]) IsSet() bool {
	switch v := any(option.Value).(type) {
	case string:
		return v != ""
	case []any:
		return len(v) > 0
	default:
		rv := reflect.ValueOf(option.Value)
		if rv.Kind() == reflect.Slice {
			return rv.Len() > 0
		}
		return false
	}
}

// IsEmpty returns true if the option is not set.
func (option DogowaiterOption[T]) IsEmpty() bool {
	return !option.IsSet()
}

func (option DogowaiterOption[T]) IsDefault() bool {
	return option.Source == DogowaiterOptionSourceDefault
}
