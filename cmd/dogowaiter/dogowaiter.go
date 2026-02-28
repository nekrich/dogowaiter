package dogowaiter

import (
	"log/slog"
	"os"

	"dogowaiter/internal/dogowaiter"
	"dogowaiter/internal/dogowaiteroptions"
)

func RunDogowaiter() {

	configuration, logger, err := dogowaiteroptions.BuildDogowaiterConfiguration()
	if err != nil {
		slog.Error("failed to build dogowaiter options", "error", err)
		os.Exit(1)
	}

	dogowaiterRunOptions := dogowaiter.RunOptions{
		Dependencies:   configuration.Dependencies,
		ConfigFilePath: configuration.ConfigFilePath,
		HealthFile:     configuration.HealthFile,
		DockerHost:     configuration.DockerHost,
		Logger:         logger,
	}

	dogowaiter.Run(&dogowaiterRunOptions)
}
