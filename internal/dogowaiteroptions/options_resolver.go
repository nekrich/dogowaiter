package dogowaiteroptions

import "strings"

// DogowaiterCommandOptions is the common set of options from CLI args or env.
// Getters use Get prefix to avoid conflict with struct fields of the same name.
type DogowaiterOptionsInterface interface {
	GetConfigFile() string
	GetDependencies() string
	GetHealthFile() string
	GetDockerHost() string
	GetLogLevel() string
}

// DogowaiterOptionsResolver merges args and env into DogowaiterOptions (args take precedence).
type DogowaiterOptionsResolver struct{}

// Resolve returns DogowaiterOptions with args overriding env for each non-empty field.
// If configFile is nil, file values are treated as empty (args/env/defaults only).
func (optionsResolver DogowaiterOptionsResolver) Resolve(arguments DogowaiterCommandArguments, env DogowaiterCommandEnv, configFile *DogowaiterConfigurationFile, defaults DogowaiterCommandDefaults) DogowaiterOptions {
	var fileDeps []DogowaiterConfigurationDependency
	var fileDockerHost, fileLogLevel string
	if configFile != nil {
		fileDeps = configFile.Dependencies
		fileDockerHost = configFile.DockerHost
		fileLogLevel = configFile.LogLevel
	}
	return DogowaiterOptions{
		ConfigFile:   resolveOption(arguments.ConfigFile, env.ConfigFile, "", defaults.ConfigFile),
		Dependencies: resolveOption(DogowaiterConfigurationDependenciesFromString(arguments.Dependencies), DogowaiterConfigurationDependenciesFromString(env.Dependencies), fileDeps, []DogowaiterConfigurationDependency{}),
		HealthFile:   resolveOption(arguments.HealthFile, env.HealthFile, "", defaults.HealthFile),
		DockerHost:   resolveOption(arguments.DockerHost, env.DockerHost, fileDockerHost, defaults.DockerHost),
		LogLevel:     resolveOption(arguments.LogLevel, env.LogLevel, fileLogLevel, defaults.LogLevel),
	}
}

// resolveOption resolves the option from the argument and environment values.
// argumentValue takes precedence over envValue, which takes precedence over fileValue, which takes precedence over defaultValue.
func resolveOption[T any](argumentValue T, envValue T, fileValue T, defaultValue T) DogowaiterOption[T] {
	if !isEmpty(argumentValue) {
		return DogowaiterOption[T]{Source: DogowaiterOptionSourceArgument, Value: argumentValue}
	}
	if !isEmpty(envValue) {
		return DogowaiterOption[T]{Source: DogowaiterOptionSourceEnvironment, Value: envValue}
	}
	if !isEmpty(fileValue) {
		return DogowaiterOption[T]{Source: DogowaiterOptionSourceConfigFile, Value: fileValue}
	}
	return DogowaiterOption[T]{Source: DogowaiterOptionSourceDefault, Value: defaultValue}
}

func isEmpty[T any](value T) bool {
	switch v := any(value).(type) {
	case string:
		return v == ""
	case []any:
		return len(v) == 0
	default:
		return false
	}
}

func DogowaiterConfigurationDependenciesFromString(dependenciesString string) []DogowaiterConfigurationDependency {
	dependencyStrings := strings.Split(dependenciesString, ",")
	dependencies := []DogowaiterConfigurationDependency{}
	for _, dependencyString := range dependencyStrings {
		dependencyParts := strings.Split(strings.TrimSpace(dependencyString), ":")

		var dependencyName string

		var dependencyStack string
		if len(dependencyParts) == 2 {
			dependencyStack = dependencyParts[0]
			dependencyName = dependencyParts[1]
		} else if len(dependencyParts) == 1 {
			dependencyName = dependencyParts[0]
		}

		if dependencyName == "" {
			continue
		}

		dependencies = append(dependencies, DogowaiterConfigurationDependency{Name: dependencyName, Stack: dependencyStack})
	}
	return dependencies
}
