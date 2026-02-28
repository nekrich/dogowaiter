package dogowaiteroptions

import "testing"

func TestDogowaiterOption_IsSet(t *testing.T) {
	tests := []struct {
		name   string
		option DogowaiterOption[string]
		want   bool
	}{
		{"empty value", DogowaiterOption[string]{Value: ""}, false},
		{"non-empty value", DogowaiterOption[string]{Value: "x"}, true},
		{"default source with value", DogowaiterOption[string]{Source: DogowaiterOptionSourceDefault, Value: "info"}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.option.IsSet(); got != tt.want {
				t.Errorf("IsSet() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDogowaiterOption_IsSet_slice(t *testing.T) {
	tests := []struct {
		name   string
		option DogowaiterOption[[]DogowaiterConfigurationDependency]
		want   bool
	}{
		{"nil slice", DogowaiterOption[[]DogowaiterConfigurationDependency]{Value: nil}, false},
		{"empty slice", DogowaiterOption[[]DogowaiterConfigurationDependency]{Value: []DogowaiterConfigurationDependency{}}, false},
		{"non-empty slice", DogowaiterOption[[]DogowaiterConfigurationDependency]{Value: []DogowaiterConfigurationDependency{{Name: "web", Stack: ""}}}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.option.IsSet(); got != tt.want {
				t.Errorf("IsSet() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDogowaiterOption_IsEmpty(t *testing.T) {
	tests := []struct {
		name   string
		option DogowaiterOption[string]
		want   bool
	}{
		{"empty value", DogowaiterOption[string]{Value: ""}, true},
		{"non-empty value", DogowaiterOption[string]{Value: "x"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.option.IsEmpty(); got != tt.want {
				t.Errorf("IsEmpty() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDogowaiterOption_IsDefault(t *testing.T) {
	tests := []struct {
		name   string
		option DogowaiterOption[string]
		want   bool
	}{
		{"default source", DogowaiterOption[string]{Source: DogowaiterOptionSourceDefault, Value: "info"}, true},
		{"environment source", DogowaiterOption[string]{Source: DogowaiterOptionSourceEnvironment, Value: "info"}, false},
		{"argument source", DogowaiterOption[string]{Source: DogowaiterOptionSourceArgument, Value: "info"}, false},
		{"config_file source", DogowaiterOption[string]{Source: DogowaiterOptionSourceConfigFile, Value: "info"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.option.IsDefault(); got != tt.want {
				t.Errorf("IsDefault() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDogowaiterOptionSource_constants(t *testing.T) {
	if DogowaiterOptionSourceDefault != "default" {
		t.Errorf("DogowaiterOptionSourceDefault = %q, want default", DogowaiterOptionSourceDefault)
	}
	if DogowaiterOptionSourceEnvironment != "environment" {
		t.Errorf("DogowaiterOptionSourceEnvironment = %q, want environment", DogowaiterOptionSourceEnvironment)
	}
	if DogowaiterOptionSourceConfigFile != "config_file" {
		t.Errorf("DogowaiterOptionSourceConfigFile = %q, want config_file", DogowaiterOptionSourceConfigFile)
	}
	if DogowaiterOptionSourceArgument != "argument" {
		t.Errorf("DogowaiterOptionSourceArgument = %q, want argument", DogowaiterOptionSourceArgument)
	}
}
