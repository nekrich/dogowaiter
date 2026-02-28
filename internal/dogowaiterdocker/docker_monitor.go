package dogowaiterdocker

import (
	"context"
	"dogowaiter/internal/dogowaiteroptions"
	"log/slog"
	"slices"

	"github.com/docker/docker/api/types/events"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
)

// DockerEventsSource provides a stream of Docker events. *client.Client implements it.
type DockerEventsSource interface {
	Events(ctx context.Context, options events.ListOptions) (<-chan events.Message, <-chan error)
	Close() error
}

type DogowaiterDockerMonitorInterface interface {
	Monitor(ctx context.Context, cli *client.Client, deps []dogowaiteroptions.DogowaiterConfigurationDependency, logger *slog.Logger) error
}

type DogowaiterDockerMonitorConfiguration struct {
	DockerHost   string
	Dependencies []dogowaiteroptions.DogowaiterConfigurationDependency
	Logger       *slog.Logger
}

type DogowaiterDockerMonitor struct {
	Configuration DogowaiterDockerMonitorConfiguration
	EventSource   DockerEventsSource // required
	cancelEvents  context.CancelFunc
}

type DogowaiterDockerEvent struct {
	ReloadConfig            bool
	UpdateHealthCheckResult bool
	ContainerID             string
}

func (monitor *DogowaiterDockerMonitor) Close() {
	monitor.setEventSource(nil)
}

func (monitor *DogowaiterDockerMonitor) setEventSource(source DockerEventsSource) {
	if monitor.EventSource != nil {
		_ = monitor.EventSource.Close()
	}
	if monitor.cancelEvents != nil {
		monitor.cancelEvents()
	}
	monitor.EventSource = source
	monitor.cancelEvents = nil
}

func (monitor *DogowaiterDockerMonitor) MonitorDependencies() chan DogowaiterDockerEvent {
	backgroundContext := context.Background()
	triggerCh := make(chan DogowaiterDockerEvent, 1)
	eventContext, cancelEvents := context.WithCancel(backgroundContext)
	monitor.cancelEvents = cancelEvents
	go monitor.subscribeEvents(eventContext, monitor.EventSource, triggerCh)
	return triggerCh
}

func (monitor *DogowaiterDockerMonitor) subscribeEvents(ctx context.Context, eventSource DockerEventsSource, triggerCh chan<- DogowaiterDockerEvent) {
	logger := monitor.Configuration.Logger
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
			}
			opts := events.ListOptions{
				Filters: newEventFilters(),
			}
			evCh, errCh := eventSource.Events(ctx, opts)
			for {
				select {
				case <-ctx.Done():
					return
				case err, ok := <-errCh:
					if ok && err != nil {
						logger.Error("events error", "error", err)
					}
					goto reconnect
				case event, ok := <-evCh:
					if !ok {
						goto reconnect
					}
					logger.Debug("event", "type", event.Type, "action", event.Action, "id", event.Actor.ID)
					if event.Type == events.ContainerEventType && isRelevantAction(string(event.Action)) {
						dockerEvent := DogowaiterDockerEvent{
							ReloadConfig:            mustReloadConfig(event.Action),
							UpdateHealthCheckResult: mustUpdateHealthCheckResult(event.Action),
							ContainerID:             event.Actor.ID,
						}
						select {
						case triggerCh <- dockerEvent:
						default:
							// already pending trigger, skip
						}
					}
				}
			}
		reconnect:
			logger.Info("events stream ended, reconnecting...")
		}
	}()
}

// mustReloadConfig returns true if the event action requires a config reload.
func mustReloadConfig(eventAction events.Action) bool {
	return slices.Contains(reloadConfigActions(), events.Action(eventAction))
}

// reloadConfigActions returns the list of actions that require a config reload.
func reloadConfigActions() []events.Action {
	return []events.Action{
		events.ActionCreate,
		events.ActionDelete,
	}
}

// mustUpdateHealthCheckResult returns true if the event action requires an update to the health check result.
func mustUpdateHealthCheckResult(eventAction events.Action) bool {
	return slices.Contains(updateHealthCheckResultActions(), events.Action(eventAction))
}

// updateHealthcheckResultActions returns the list of actions that require an update to the health check result.
func updateHealthCheckResultActions() []events.Action {
	return []events.Action{
		events.ActionStart,
		events.ActionDie,
		events.ActionHealthStatus,
	}
}

// relevantActions returns the list of actions that are relevant to dogowaiter.
func relevantActions() []events.Action {
	return slices.Concat(reloadConfigActions(), updateHealthCheckResultActions())
}

// newEventFilters returns Docker event filters for type=container and event in (start, die, health_status, create, delete).
func newEventFilters() filters.Args {
	args := filters.NewArgs()
	args.Add("type", string(events.ContainerEventType))
	for _, action := range relevantActions() {
		args.Add("event", string(action))
	}
	return args
}

// isRelevantAction returns true if action is relevant to dogowaiter.
func isRelevantAction(action string) bool {
	return slices.Contains(relevantActions(), events.Action(action))
}
