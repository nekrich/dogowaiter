package dogowaiteroptions

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDogowaiterConfigurationDependenciesFromString(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		output []DogowaiterConfigurationDependency
	}{
		{"empty", "", nil},
		{"single service", "web", []DogowaiterConfigurationDependency{{Name: "web", Stack: ""}}},
		{"stack:service", "mystack:web", []DogowaiterConfigurationDependency{{Name: "web", Stack: "mystack"}}},
		{"comma-separated", "web,worker", []DogowaiterConfigurationDependency{
			{Name: "web", Stack: ""},
			{Name: "worker", Stack: ""},
		}},
		{"with spaces", " stack:svc , other ", []DogowaiterConfigurationDependency{
			{Name: "svc", Stack: "stack"},
			{Name: "other", Stack: ""},
		}},
		{"skip empty segment", "web,,worker", []DogowaiterConfigurationDependency{
			{Name: "web", Stack: ""},
			{Name: "worker", Stack: ""},
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DogowaiterConfigurationDependenciesFromString(tt.input)
			if len(got) != len(tt.output) {
				t.Errorf("len = %d, want %d; got %+v", len(got), len(tt.output), got)
				return
			}
			for i := range got {
				if !got[i].Equal(tt.output[i]) {
					t.Errorf("at %d: got %+v, want %+v", i, got[i], tt.output[i])
				}
			}
		})
	}
}

func TestDogowaiterConfigurationDependency_Equal(t *testing.T) {
	a := DogowaiterConfigurationDependency{Name: "web", Stack: "s1"}
	b := DogowaiterConfigurationDependency{Name: "web", Stack: "s1"}
	c := DogowaiterConfigurationDependency{Name: "web", Stack: ""}
	d := DogowaiterConfigurationDependency{Name: "other", Stack: "s1"}

	if !a.Equal(b) {
		t.Error("a.Equal(b) want true")
	}
	if a.Equal(c) {
		t.Error("a.Equal(c) want false (different stack)")
	}
	if a.Equal(d) {
		t.Error("a.Equal(d) want false (different name)")
	}
}

func TestCheckConfigFileExists(t *testing.T) {
	dir := t.TempDir()
	existing := filepath.Join(dir, "exists.yaml")
	if err := os.WriteFile(existing, []byte("x"), 0644); err != nil {
		t.Fatal(err)
	}

	missing := filepath.Join(dir, "missing.yaml")

	tests := []struct {
		path       string
		required   bool
		wantExists bool
		wantErr    bool
	}{
		{existing, false, true, false},
		{existing, true, true, false},
		{missing, false, false, false},
		{missing, true, false, true},
		{"", false, false, false},
		{"", true, false, true},
	}
	for _, tt := range tests {
		exists, err := checkConfigFileExists(tt.path, tt.required)
		if exists != tt.wantExists {
			t.Errorf("checkConfigFileExists(%q, %v) exists = %v, want %v", tt.path, tt.required, exists, tt.wantExists)
		}
		if (err != nil) != tt.wantErr {
			t.Errorf("checkConfigFileExists(%q, %v) err = %v, wantErr %v", tt.path, tt.required, err, tt.wantErr)
		}
	}
}

func TestMergeOptionValue(t *testing.T) {
	tests := []struct {
		name        string
		optionValue string
		resolved    DogowaiterOption[string]
		want        string
	}{
		{"use config when resolved is default", "file-value", DogowaiterOption[string]{Source: DogowaiterOptionSourceDefault, Value: "default"}, "file-value"},
		{"use resolved when non-default", "file-value", DogowaiterOption[string]{Source: DogowaiterOptionSourceArgument, Value: "arg"}, "arg"},
		{"use resolved when option empty", "", DogowaiterOption[string]{Source: DogowaiterOptionSourceDefault, Value: "default"}, "default"},
		{"use resolved when option empty and env", "", DogowaiterOption[string]{Source: DogowaiterOptionSourceEnvironment, Value: "env"}, "env"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := mergeOptionValue("test", tt.optionValue, tt.resolved)
			if got != tt.want {
				t.Errorf("mergeOptionValue() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestEffectiveConfigFilePath(t *testing.T) {
	defaults := NewDogowaiterCommandDefaults()
	t.Run("args override", func(t *testing.T) {
		args := DogowaiterCommandArguments{ConfigFile: "/arg.yaml"}
		env := DogowaiterCommandEnv{}
		got := effectiveConfigFilePath(args, env, defaults)
		if got != "/arg.yaml" {
			t.Errorf("effectiveConfigFilePath() = %q", got)
		}
	})
	t.Run("env when args empty", func(t *testing.T) {
		args := DogowaiterCommandArguments{}
		env := DogowaiterCommandEnv{ConfigFile: "/env.yaml"}
		got := effectiveConfigFilePath(args, env, defaults)
		if got != "/env.yaml" {
			t.Errorf("effectiveConfigFilePath() = %q", got)
		}
	})
	t.Run("default when args and env empty", func(t *testing.T) {
		args := DogowaiterCommandArguments{}
		env := DogowaiterCommandEnv{}
		got := effectiveConfigFilePath(args, env, defaults)
		if got != defaults.ConfigFile {
			t.Errorf("effectiveConfigFilePath() = %q, want %q", got, defaults.ConfigFile)
		}
	})
}

func TestBuildDogowaiterConfigurationWithArguments(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "dogowaiter.yaml")
	if err := os.WriteFile(cfgPath, []byte("depends_on:\n  - name: fromfile\nlog_level: info\n"), 0644); err != nil {
		t.Fatal(err)
	}

	args := DogowaiterCommandArguments{
		ConfigFile:   cfgPath,
		Dependencies: "cli:svc",
		HealthFile:   "/tmp/custom-health",
	}

	cfg, err := buildDogowaiterConfigurationWithArguments(args)
	if err != nil {
		t.Fatal(err)
	}
	// Dependencies from CLI override file
	if len(cfg.Dependencies) != 1 {
		t.Fatalf("Dependencies len = %d", len(cfg.Dependencies))
	}
	if cfg.Dependencies[0].Name != "svc" || cfg.Dependencies[0].Stack != "cli" {
		t.Errorf("Dependencies[0] = %+v", cfg.Dependencies[0])
	}
	if cfg.HealthFile != "/tmp/custom-health" {
		t.Errorf("HealthFile = %q", cfg.HealthFile)
	}
}
