package dogowaiterconfigfilemonitor

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/fsnotify/fsnotify"
)

type DogowaiterConfigFileMonitorInterface interface {
	Monitor() chan<- struct{}
}

type DogowaiterConfigFileMonitorConfiguration struct {
	ConfigFilePath   string
	DebounceInterval time.Duration // zero = use defaultDebounceInterval
}

type DogowaiterConfigFileMonitor struct {
	Configuration  DogowaiterConfigFileMonitorConfiguration
	Logger         *slog.Logger
	symlinkMonitor *DogowaiterConfigFileMonitor
	watcher        *fsnotify.Watcher
}

func (monitor DogowaiterConfigFileMonitor) Close() {
	monitor.setSymlinkMonitor(nil)
	monitor.setWatcher(nil)
}

func (monitor DogowaiterConfigFileMonitor) setWatcher(watcher *fsnotify.Watcher) {
	if monitor.watcher != nil {
		monitor.watcher.Close()
	}
	monitor.watcher = watcher
}

func (monitor DogowaiterConfigFileMonitor) setSymlinkMonitor(symlinkMonitor *DogowaiterConfigFileMonitor) {
	if monitor.symlinkMonitor != nil {
		monitor.symlinkMonitor.Close()
	}
	monitor.symlinkMonitor = symlinkMonitor
}

func (monitor DogowaiterConfigFileMonitor) Monitor(configFileUpdateChannel chan<- struct{}) error {
	return monitor.monitor(monitor.Configuration.ConfigFilePath, configFileUpdateChannel, true)
}

func (monitor *DogowaiterConfigFileMonitor) monitor(file string, configFileUpdateChannel chan<- struct{}, resolveSymlink bool) error {
	if file == "" {
		return nil
	}
	if _, err := os.Stat(file); os.IsNotExist(err) {
		return nil
	}
	reloadCh := make(chan struct{}, 1)
	removedCh := make(chan struct{}, 1)

	debounce := monitor.Configuration.DebounceInterval
	if debounce == 0 {
		debounce = defaultDebounceInterval
	}
	watcher, watchErr := watchFile(file, reloadCh, removedCh, debounce, monitor.Logger)
	if watchErr != nil {
		monitor.Logger.Error("failed to watch config file", "error", watchErr)
		return watchErr
	}
	monitor.setWatcher(watcher)

	mergeChannels(context.Background(), configFileUpdateChannel, monitor.Logger, reloadCh, removedCh)

	if resolveSymlink {
		symlinkMonitor, monitorErr := monitor.monitorSymlink(file, configFileUpdateChannel)
		if monitorErr != nil {
			return monitorErr
		}
		monitor.setSymlinkMonitor(symlinkMonitor)
	}

	return nil
}

func (monitor DogowaiterConfigFileMonitor) monitorSymlink(file string, configFileUpdateChannel chan<- struct{}) (*DogowaiterConfigFileMonitor, error) {
	resolvedConfigFilePath, err := filepath.EvalSymlinks(file)
	if err != nil {
		return nil, err
	}
	if resolvedConfigFilePath == monitor.Configuration.ConfigFilePath {
		return nil, nil
	}

	symlinkMonitor := DogowaiterConfigFileMonitor{
		Configuration: DogowaiterConfigFileMonitorConfiguration{
			ConfigFilePath:   resolvedConfigFilePath,
			DebounceInterval: monitor.Configuration.DebounceInterval,
		},
		Logger: monitor.Logger,
	}
	monitorErr := symlinkMonitor.monitor(resolvedConfigFilePath, configFileUpdateChannel, false)
	if monitorErr != nil {
		return nil, monitorErr
	}
	return &symlinkMonitor, nil
}

const defaultDebounceInterval = 3 * time.Second

// WatchConfig watches the resolved config file path for changes. On Write or Create it debounces
// then sends on reloadCh. On Remove it sends on removedCh. Caller must Close the returned
// watcher when done. reloadCh and removedCh must be buffered (e.g. size 1); sends are non-blocking.
func watchFile(filePath string, reloadCh chan<- struct{}, removedCh chan<- struct{}, debounceInterval time.Duration, logger *slog.Logger) (*fsnotify.Watcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	if err := watcher.Add(filePath); err != nil {
		_ = watcher.Close()
		return nil, err
	}
	filePath = filepath.Clean(filePath)

	backgroundContext := context.Background()
	go watchFileBackground(backgroundContext, watcher, filePath, reloadCh, removedCh, debounceInterval, logger)
	return watcher, nil
}

func mergeChannels(ctx context.Context, mergedChannel chan<- struct{}, logger *slog.Logger, channels ...chan struct{}) {
	go func() {
		for _, channel := range channels {
			go func(channel chan struct{}) {
				for {
					select {
					case <-ctx.Done():
						return
					case value := <-channel:
						mergedChannel <- value
					}
				}
			}(channel)
		}
	}()
}

func watchFileBackground(ctx context.Context, watcher *fsnotify.Watcher, filePath string, reloadCh chan<- struct{}, removedCh chan<- struct{}, debounceInterval time.Duration, logger *slog.Logger) {
	go func() {
		var debounceTimer *time.Timer
		for {
			select {
			case <-ctx.Done():
				return
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				// Normalize for comparison (fsnotify may use OS path form).
				evPath := filepath.Clean(event.Name)
				if evPath != filePath {
					continue
				}
				switch {
				case event.Has(fsnotify.Remove):
					if debounceTimer != nil {
						debounceTimer.Stop()
						debounceTimer = nil
					}
					select {
					case removedCh <- struct{}{}:
					default:
					}
				case event.Has(fsnotify.Write) || event.Has(fsnotify.Create):
					if debounceTimer != nil {
						debounceTimer.Stop()
						debounceTimer = nil
					}
					debounceTimer = time.AfterFunc(debounceInterval, func() {
						select {
						case reloadCh <- struct{}{}:
						default:
						}
					})
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				logger.Error("config watch error", "error", err)
			}
		}
	}()
}
