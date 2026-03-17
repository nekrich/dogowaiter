package dogowaiterconfigfilemonitor

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// Short debounce for tests that wait for reload so they run in ~100ms instead of 3s.
const testDebounce = 20 * time.Millisecond

func TestMergeChannels_forwardsToMerged(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ch1 := make(chan struct{}, 1)
	ch2 := make(chan struct{}, 1)
	merged := make(chan struct{}, 2)
	logger := slog.Default()

	mergeChannels(ctx, merged, logger, ch1, ch2)

	ch1 <- struct{}{}
	select {
	case <-merged:
	case <-time.After(time.Second):
		t.Fatal("expected value from merged channel")
	}

	ch2 <- struct{}{}
	select {
	case <-merged:
	case <-time.After(time.Second):
		t.Fatal("expected value from merged channel")
	}

	cancel()
	time.Sleep(50 * time.Millisecond) // let goroutines exit
}

func TestWatchFile_validPath_returnsWatcher(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(path, []byte("x"), 0644); err != nil {
		t.Fatal(err)
	}

	reloadCh := make(chan struct{}, 1)
	removedCh := make(chan struct{}, 1)
	logger := slog.Default()

	watcher, err := watchFile(path, reloadCh, removedCh, defaultDebounceInterval, logger)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = watcher.Close() }()

	if watcher == nil {
		t.Fatal("watcher is nil")
	}
}

func TestWatchFile_missingPath_returnsError(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nonexistent.yaml")
	reloadCh := make(chan struct{}, 1)
	removedCh := make(chan struct{}, 1)
	logger := slog.Default()

	watcher, err := watchFile(path, reloadCh, removedCh, defaultDebounceInterval, logger)
	if err == nil {
		if watcher != nil {
			_ = watcher.Close()
		}
		t.Fatal("expected error when watching missing path")
	}
	if watcher != nil {
		t.Error("watcher should be nil on error")
	}
}

func TestWatchFile_writeSendsReloadAfterDebounce(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(path, []byte("initial"), 0644); err != nil {
		t.Fatal(err)
	}

	reloadCh := make(chan struct{}, 1)
	removedCh := make(chan struct{}, 1)
	logger := slog.Default()

	watcher, err := watchFile(path, reloadCh, removedCh, testDebounce, logger)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = watcher.Close() }()

	if err := os.WriteFile(path, []byte("updated"), 0644); err != nil {
		t.Fatal(err)
	}

	select {
	case <-reloadCh:
	case <-time.After(testDebounce + 100*time.Millisecond):
		t.Fatal("expected reload signal after write (debounce)")
	}
}

func TestMonitorSymlink_regularFile_samePath_returnsNilMonitor(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(path, []byte("x"), 0644); err != nil {
		t.Fatal(err)
	}

	resolved, err := filepath.EvalSymlinks(path)
	if err != nil {
		t.Fatal(err)
	}

	monitor := DogowaiterConfigFileMonitor{
		Configuration: DogowaiterConfigFileMonitorConfiguration{ConfigFilePath: resolved},
		Logger:        slog.Default(),
	}
	ch := make(chan struct{}, 1)
	symlinkMonitor, err := monitor.monitorSymlink(path, ch)
	if err != nil {
		t.Fatal(err)
	}
	if symlinkMonitor != nil {
		t.Error("expected nil symlinkMonitor when path equals resolved path")
		symlinkMonitor.Close()
	}
}

func TestMonitorSymlink_symlinkTarget_returnsMonitor(t *testing.T) {
	dir := t.TempDir()
	realPath := filepath.Join(dir, "real.yaml")
	linkPath := filepath.Join(dir, "link.yaml")
	if err := os.WriteFile(realPath, []byte("x"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(realPath, linkPath); err != nil {
		t.Skip("symlink not supported on this system")
	}

	monitor := DogowaiterConfigFileMonitor{
		Configuration: DogowaiterConfigFileMonitorConfiguration{ConfigFilePath: linkPath},
		Logger:        slog.Default(),
	}
	ch := make(chan struct{}, 1)
	symlinkMonitor, err := monitor.monitorSymlink(linkPath, ch)
	if err != nil {
		t.Fatal(err)
	}
	if symlinkMonitor == nil {
		t.Fatal("expected non-nil symlinkMonitor for symlink path")
	}
	symlinkMonitor.Close()
}

func TestDogowaiterConfigFileMonitor_Close_noPanic(t *testing.T) {
	m := DogowaiterConfigFileMonitor{}
	m.Close()
}

func TestDogowaiterConfigFileMonitor_Monitor_and_Close(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(path, []byte("initial"), 0644); err != nil {
		t.Fatal(err)
	}

	ch := make(chan struct{}, 1)
	monitor := DogowaiterConfigFileMonitor{
		Configuration: DogowaiterConfigFileMonitorConfiguration{
			ConfigFilePath:   path,
			DebounceInterval: testDebounce,
		},
		Logger: slog.Default(),
	}

	err := monitor.Monitor(ch)
	if err != nil {
		t.Fatal(err)
	}
	defer monitor.Close()

	if err := os.WriteFile(path, []byte("updated"), 0644); err != nil {
		t.Fatal(err)
	}

	select {
	case <-ch:
	case <-time.After(testDebounce + 100*time.Millisecond):
		t.Fatal("expected config update signal after file write")
	}
}
