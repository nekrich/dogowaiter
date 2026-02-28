package dogowaiteroptions

import (
	"log/slog"
	"os"
	"strings"

	"github.com/jessevdk/go-flags"
)

// DogowaiterConfiguration is the resolved configuration (from file and options merge). File content is loaded via DogowaiterConfigurationFile.
type DogowaiterConfiguration struct {
	Dependencies   []DogowaiterConfigurationDependency
	HealthFile     string
	DockerHost     string
	LogLevel       string
	ConfigFilePath string
}

func resolveOptions(commandArguments DogowaiterCommandArguments) DogowaiterOptions {
	commandEnv := NewDogowaiterCommandEnv()
	commandDefaults := NewDogowaiterCommandDefaults()
	configFilePath := effectiveConfigFilePath(commandArguments, commandEnv, commandDefaults)
	strictConfigCheck := commandArguments.ConfigFile != "" || commandEnv.ConfigFile != ""

	configFile, err := loadDogowaiterConfigurationFile(configFilePath, strictConfigCheck)
	if err != nil {
		slog.Default().Error("failed to load config file", "error", err)
		os.Exit(1)
	}

	return new(DogowaiterOptionsResolver).Resolve(commandArguments, commandEnv, configFile, commandDefaults)
}

func effectiveConfigFilePath(args DogowaiterCommandArguments, env DogowaiterCommandEnv, defaults DogowaiterCommandDefaults) string {
	if args.ConfigFile != "" {
		return args.ConfigFile
	}
	if env.ConfigFile != "" {
		return env.ConfigFile
	}
	return defaults.ConfigFile
}

func buildDogowaiterConfigurationWithArguments(commandArguments DogowaiterCommandArguments) (*DogowaiterConfiguration, error) {
	resolvedOptions := resolveOptions(commandArguments)
	return &DogowaiterConfiguration{
		ConfigFilePath: resolvedOptions.ConfigFile.Value,
		Dependencies:   resolvedOptions.Dependencies.Value,
		HealthFile:     resolvedOptions.HealthFile.Value,
		DockerHost:     resolvedOptions.DockerHost.Value,
		LogLevel:       resolvedOptions.LogLevel.Value,
	}, nil
}

func mergeOptionValue[T any](optionName string, optionValue T, resolvedOption DogowaiterOption[T]) T {
	if isEmpty(optionValue) || !resolvedOption.IsDefault() {
		return resolvedOption.Value
	}
	return optionValue
}

func BuildDogowaiterConfiguration() (*DogowaiterConfiguration, *slog.Logger, error) {
	var commandArguments DogowaiterCommandArguments
	_, err := flags.Parse(&commandArguments)
	if err != nil {
		if flagsErr, ok := err.(*flags.Error); ok && flagsErr.Type == flags.ErrHelp {
			return nil, nil, err
		}
		os.Exit(1)
	}

	configuration, err := buildDogowaiterConfigurationWithArguments(commandArguments)
	if err != nil {
		slog.Error("failed to build dogowaiter configuration", "error", err)
		os.Exit(1)
	}

	logger := newLogger(configuration.LogLevel)

	return configuration, logger, nil
}

func newLogger(level string) *slog.Logger {
	switch strings.ToLower(level) {
	case "debug":
		return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	case "info":
		return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	case "warn":
		return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelWarn}))
	case "warning":
		return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelWarn}))
	case "error":
		return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	}
	return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
}
