package dogowaiter

import (
	"context"
	"dogowaiter/internal/dogowaiterconfigfilemonitor"
	"dogowaiter/internal/dogowaiterdocker"
	"dogowaiter/internal/dogowaiterhealthfile"
	"dogowaiter/internal/dogowaiteroptions"
	"log/slog"
	"os"
	"os/signal"
	"slices"
	"syscall"

	"github.com/docker/docker/client"
)

// RunOptions are optional overrides for Run (e.g. from CLI). Nil means use env only.
type RunOptions struct {
	Dependencies   []dogowaiteroptions.DogowaiterConfigurationDependency
	ConfigFilePath string
	HealthFile     string
	DockerHost     string
	Logger         *slog.Logger
}

// Run loads config (file or DOGOWAITER_DEPENDENCIES env), runs an initial dependency check to build in-memory state,
// subscribes to Docker events, and updates state only when a monitored container has a status-change event.
// When using a config file, watches it (resolved path) and reloads on change (30s debounce); if the file is removed,
// resolves path once more and exits with error status if missing. Pass opts from CLI; nil means use env only.
func Run(opts *RunOptions) {
	opts.Logger.Info("running dogowaiter", "dependencies", opts.Dependencies, "configFilePath", opts.ConfigFilePath, "healthFile", opts.HealthFile, "dockerHost", opts.DockerHost)
	logger := opts.Logger

	cli, err := client.NewClientWithOpts(client.WithHost(opts.DockerHost), client.WithAPIVersionNegotiation())
	if err != nil {
		logger.Error("failed to create docker client", "error", err)
		os.Exit(1)
	}

	ctx := context.Background()

	logger.Info("resolving dependencies", "dependencies", opts.Dependencies)
	containerResolver := dogowaiterdocker.DogowaiterDockerContainerResolver{}

	dependentContainersInStack, err := containerResolver.FindDependentDockerContainersInStack(ctx, cli, logger)
	if err != nil {
		logger.Error("failed to resolve dependent docker containers", "error", err)
		os.Exit(1)
	}
	logger.Info("dependent docker containers resolved", "dependentContainersInStack", dependentContainersInStack)
	logger.Info("dependent dependencies count", "dependentContainersInStackCount", len(dependentContainersInStack))

	allDependencies := slices.Clone(opts.Dependencies)
	for _, dependentContainerInStack := range dependentContainersInStack {
		if !slices.ContainsFunc(allDependencies, func(lhs dogowaiteroptions.DogowaiterConfigurationDependency) bool {
			return lhs.Equal(dependentContainerInStack)
		}) {
			allDependencies = append(allDependencies, dependentContainerInStack)
		}
	}
	logger.Info("all dependencies", "allDependenciesCount", len(allDependencies), "allDependencies", allDependencies)

	dockerContainers, err := containerResolver.FindRequiredDockerContainers(ctx, cli, allDependencies, logger)
	if err != nil {
		logger.Error("failed to resolve dependencies", "error", err)
		os.Exit(1)
	}
	logger.Info("docker containers resolved before initial check", "containers", dockerContainers)

	checker := dogowaiterdocker.DogowaiterDockerContainerChecker{
		DockerClient: cli,
		Logger:       logger,
	}
	containers := runInitialCheck(ctx, checker, allDependencies, dockerContainers, logger)
	logger.Info("docker containers resolved after initial check", "containers", dockerContainers)
	state := buildState(containers)

	healthFile := dogowaiterhealthfile.DogowaiterHealthFile{
		FilePath: opts.HealthFile,
	}

	if err := healthFile.Write(&state); err != nil {
		logger.Error("failed to write health file", "error", err)
		os.Exit(1)
	}
	logRoundFromResult(&state, logger)

	dockerMonitor := &dogowaiterdocker.DogowaiterDockerMonitor{
		Configuration: dogowaiterdocker.DogowaiterDockerMonitorConfiguration{
			DockerHost:   opts.DockerHost,
			Dependencies: allDependencies,
			Logger:       logger,
		},
		EventSource: cli,
	}
	dockerTriggerCh := dockerMonitor.MonitorDependencies()
	defer dockerMonitor.Close()

	configFileMonitor := &dogowaiterconfigfilemonitor.DogowaiterConfigFileMonitor{
		Configuration: dogowaiterconfigfilemonitor.DogowaiterConfigFileMonitorConfiguration{
			ConfigFilePath: opts.ConfigFilePath,
		},
		Logger: logger,
	}
	configFileTriggerCh := make(chan struct{}, 1)
	if err := configFileMonitor.Monitor(configFileTriggerCh); err != nil {
		logger.Error("failed to start config file monitor", "error", err)
		os.Exit(1)
	}
	defer configFileMonitor.Close()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT)

	for {
		select {
		case <-sigCh:
			_ = healthFile.Remove()
			logger.Info("shutting down")
			os.Exit(0)
		case <-configFileTriggerCh:
			reloadConfig(ctx, cli, &healthFile, opts)
		case dockerEvent := <-dockerTriggerCh:
			if dockerEvent.ReloadConfig {
				reloadConfig(ctx, cli, &healthFile, opts)
			} else if dockerEvent.UpdateHealthCheckResult && dockerEvent.ContainerID != "" {
				updated := updateContainerState(ctx, checker, &state, dockerEvent.ContainerID)
				if updated {
					if err := healthFile.Write(&state); err != nil {
						logger.Error("failed to write health file", "error", err)
					}
					logRoundFromResult(&state, logger)
				}
			}
		}
	}
}

// reloadConfig reloads config from configPath, re-resolves, re-runs initial check, and replaces state and cfg. On load or resolve failure logs and keeps current state/cfg.
func reloadConfig(ctx context.Context, cli *client.Client, healthFile *dogowaiterhealthfile.DogowaiterHealthFile, opts *RunOptions) {
	newConfiguration, logger, err := dogowaiteroptions.BuildDogowaiterConfiguration()
	if err != nil {
		logger.Error("failed to reload config", "error", err)
		return
	}
	dockerContainers, err := (dogowaiterdocker.DogowaiterDockerContainerResolver{}).FindRequiredDockerContainers(ctx, cli, newConfiguration.Dependencies, logger)
	if err != nil {
		logger.Error("failed to reload config", "error", err)
		return
	}
	checker := dogowaiterdocker.DogowaiterDockerContainerChecker{
		DockerClient: cli,
		Logger:       logger,
	}
	containers := runInitialCheck(ctx, checker, newConfiguration.Dependencies, dockerContainers, logger)
	newState := buildState(containers)
	newRunOptions := RunOptions{
		Dependencies:   newConfiguration.Dependencies,
		ConfigFilePath: newConfiguration.ConfigFilePath,
		HealthFile:     newConfiguration.HealthFile,
		DockerHost:     newConfiguration.DockerHost,
		Logger:         logger,
	}
	if err := healthFile.Write(&newState); err != nil {
		logger.Error("failed to write health file", "error", err)
	}
	*opts = newRunOptions
	logRoundFromResult(&newState, logger)
}

// buildState builds HealthCheckResult from the full container list. Healthy is true when all entries have IsReady.
func buildState(containers []dogowaiterhealthfile.HealthContainer) dogowaiterhealthfile.HealthCheckResult {
	allReady := true
	for i := range containers {
		if !containers[i].IsReady {
			allReady = false
			break
		}
	}
	return dogowaiterhealthfile.BuildHealthCheckResult(allReady, containers)
}

// runInitialCheck runs one check per resolved container and returns the full list of HealthContainer.
func runInitialCheck(ctx context.Context, checker dogowaiterdocker.DogowaiterDockerContainerCheckerInterface, deps []dogowaiteroptions.DogowaiterConfigurationDependency, results []dogowaiterdocker.DockerContainerResolveResult, logger *slog.Logger) []dogowaiterhealthfile.HealthContainer {
	var list []dogowaiterhealthfile.HealthContainer
	for i, res := range results {
		_ = deps[i]
		if len(res.ContainerIDs) == 0 {
			list = append(list, dogowaiterhealthfile.HealthContainer{
				ContainerID: "",
				Container:   res.DependencyName,
				Reason:      dogowaiterdocker.ReasonNotStarted,
				IsReady:     false,
			})
			logger.Info("dependency", res.DependencyName, ": no container found")
			continue
		}
		for _, cid := range res.ContainerIDs {
			result, err := checker.CheckContainer(ctx, cid)
			name := result.ContainerName
			if name == "" {
				name = cid
				if len(cid) > 12 {
					name = cid[:12]
				}
			}
			reason := result.Reason
			if err != nil {
				logger.Error("failed to check container", "dependency", res.DependencyName, "container", cid[:12], "error", err)
				reason = dogowaiterdocker.ReasonNotStarted
			}
			if result.Pass {
				logger.Info("dependency", res.DependencyName, ": container", name, ": pass")
			} else {
				logger.Error("dependency", res.DependencyName, ": container", name, ": fail", "reason", reason)
			}
			list = append(list, dogowaiterhealthfile.HealthContainer{
				ContainerID: cid,
				Container:   name,
				Reason:      reason,
				IsReady:     result.Pass,
			})
		}
	}
	return list
}

// updateContainerState finds the container by ID in state, runs CheckContainer once, updates that entry, recomputes Healthy. Returns true if state changed.
func updateContainerState(ctx context.Context, checker dogowaiterdocker.DogowaiterDockerContainerCheckerInterface, state *dogowaiterhealthfile.HealthCheckResult, containerID string) bool {
	idx := -1
	containers := state.GetContainers()
	for i := range containers {
		if containers[i].ContainerID == containerID {
			idx = i
			break
		}
	}
	if idx < 0 {
		return false
	}
	entry := &containers[idx]
	result, err := checker.CheckContainer(ctx, containerID)
	if err != nil {
		entry.IsReady = false
		entry.Reason = dogowaiterdocker.ReasonNotStarted
	} else {
		entry.IsReady = result.Pass
		entry.Reason = result.Reason
		if result.ContainerName != "" {
			entry.Container = result.ContainerName
		}
	}
	previousState := cloneState(state)

	atLeastOneIsUnhealthy := slices.ContainsFunc(containers, func(container dogowaiterhealthfile.HealthContainer) bool {
		return !container.IsReady
	})

	isHealthy := len(containers) > 0 && !atLeastOneIsUnhealthy

	currentState := dogowaiterhealthfile.BuildHealthCheckResult(isHealthy, containers)

	defer func() {
		*state = currentState
	}()

	return !previousState.Equal(&currentState)
}

func cloneState(s *dogowaiterhealthfile.HealthCheckResult) *dogowaiterhealthfile.HealthCheckResult {
	if s == nil {
		return nil
	}
	healthCheckResult := dogowaiterhealthfile.BuildHealthCheckResult(s.IsHealthy(), s.GetContainers())
	return &healthCheckResult
}

// logRoundFromResult logs the overall result from state.
func logRoundFromResult(state *dogowaiterhealthfile.HealthCheckResult, logger *slog.Logger) {
	if state.IsHealthy() {
		logger.Info("round result: healthy")
		return
	}
	var failing []dogowaiterhealthfile.HealthContainer
	containers := state.GetContainers()
	for i := range containers {
		if !containers[i].IsReady {
			failing = append(failing, containers[i])
		}
	}
	logger.Info("round result: unhealthy", "failing_count", len(failing), "containers", failing)
}
