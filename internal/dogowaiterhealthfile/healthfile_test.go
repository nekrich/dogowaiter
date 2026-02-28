package dogowaiterhealthfile

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestDogowaiterHealthFile_Write_writesFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "health.json")
	healthFile := DogowaiterHealthFile{FilePath: path}

	result := BuildHealthCheckResult(true, []HealthContainer{
		{ContainerID: "id1", Container: "svc1", Reason: "running", IsReady: true},
	})

	err := healthFile.Write(&result)
	if err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var payload struct {
		Healthy    bool `json:"healthy"`
		Containers []struct {
			Container string `json:"container"`
			IsReady   bool   `json:"is_ready"`
			Reason    string `json:"reason"`
		} `json:"containers"`
	}
	if err := json.Unmarshal(data, &payload); err != nil {
		t.Fatal(err)
	}
	if !payload.Healthy {
		t.Error("payload.healthy want true")
	}
	if len(payload.Containers) != 1 {
		t.Fatalf("payload.containers len = %d, want 1", len(payload.Containers))
	}
	if payload.Containers[0].Container != "svc1" || !payload.Containers[0].IsReady || payload.Containers[0].Reason != "running" {
		t.Errorf("payload.containers[0] = %+v", payload.Containers[0])
	}
}

func TestDogowaiterHealthFile_Write_skipsWhenEqual(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "health.json")
	healthFile := DogowaiterHealthFile{FilePath: path}

	result := BuildHealthCheckResult(true, []HealthContainer{
		{ContainerID: "id1", Container: "svc1", Reason: "ok", IsReady: true},
	})

	if err := healthFile.Write(&result); err != nil {
		t.Fatal(err)
	}
	fileStatsBeforeWrite, _ := os.Stat(path)

	// same result: should not rewrite (no error, file unchanged)
	if err := healthFile.Write(&result); err != nil {
		t.Fatal(err)
	}
	fileStatsAfterWrite, _ := os.Stat(path)
	if !fileStatsBeforeWrite.ModTime().Equal(fileStatsAfterWrite.ModTime()) {
		t.Error("Write with equal result should not update file modification time")
	}
}

func TestDogowaiterHealthFile_Write_updatesWhenDifferent(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "health.json")
	healthFile := DogowaiterHealthFile{FilePath: path}

	result1 := BuildHealthCheckResult(true, []HealthContainer{
		{ContainerID: "id1", Container: "svc1", Reason: "ok", IsReady: true},
	})
	result2 := BuildHealthCheckResult(false, []HealthContainer{
		{ContainerID: "id1", Container: "svc1", Reason: "unhealthy", IsReady: false},
	})

	if err := healthFile.Write(&result1); err != nil {
		t.Fatal(err)
	}
	if err := healthFile.Write(&result2); err != nil {
		t.Fatal(err)
	}

	data, _ := os.ReadFile(path)
	var payload struct {
		Healthy bool `json:"healthy"`
	}
	_ = json.Unmarshal(data, &payload)
	if payload.Healthy {
		t.Error("payload.healthy want false after second Write")
	}
}

func TestDogowaiterHealthFile_Remove(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "health.json")
	healthFile := DogowaiterHealthFile{FilePath: path}

	result := BuildHealthCheckResult(true, nil)
	if err := healthFile.Write(&result); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatal("file should exist after Write")
	}

	if err := healthFile.Remove(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Error("file should not exist after Remove")
	}
}

func TestDogowaiterHealthFile_Remove_missingFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nonexistent.json")
	healthFile := DogowaiterHealthFile{FilePath: path}
	// os.Remove on missing file returns error in Go
	err := healthFile.Remove()
	if err != nil && !os.IsNotExist(err) {
		t.Errorf("Remove() on missing file: err = %v", err)
	}
}
