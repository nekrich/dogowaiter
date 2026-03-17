package dogowaiteroptions

import (
	"fmt"
	"log/slog"
	"os"
	"strings"

	yaml "go.yaml.in/yaml/v3"
)

// DogowaiterConfigDependency is one dependency to monitor. Name is the service name; Stack is optional (empty = any stack).
type DogowaiterConfigurationDependency struct {
	Name  string `yaml:"name"`
	Stack string `yaml:"stack"`
}

func (lhs DogowaiterConfigurationDependency) Equal(rhs DogowaiterConfigurationDependency) bool {
	return strings.Compare(lhs.Name, rhs.Name) == 0 && strings.Compare(lhs.Stack, rhs.Stack) == 0
}

// DogowaiterConfigurationFile is the resolved configuration from the config file.
type DogowaiterConfigurationFile struct {
	Dependencies []DogowaiterConfigurationDependency `yaml:"depends_on"`
	DockerHost   string                              `yaml:"docker_host,omitempty"`
	LogLevel     string                              `yaml:"log_level,omitempty"`
}

func checkConfigFileExists(configFilePath string, fileIsRequired bool) (bool, error) {
	if configFilePath == "" {
		if fileIsRequired {
			return false, fmt.Errorf("config file path is empty")
		}
		return false, nil
	}

	if _, err := os.Stat(configFilePath); os.IsNotExist(err) {
		if fileIsRequired {
			return false, fmt.Errorf("checkConfigFileExists Config file does not exist at path: %s", configFilePath)
		}
		return false, nil
	}

	return true, nil
}

func loadDogowaiterConfigurationFile(configFilePath string, fileIsRequired bool) (*DogowaiterConfigurationFile, error) {
	if exists, err := checkConfigFileExists(configFilePath, fileIsRequired); err != nil {
		slog.Error("config file does not exist", "configFilePath", configFilePath, "error", err)
		return nil, err
	} else if !exists {
		if fileIsRequired {
			return nil, fmt.Errorf("config file does not exist at path: %s", configFilePath)
		}
		return nil, nil
	}

	configurationFile := DogowaiterConfigurationFile{}

	// read the config file
	data, err := os.ReadFile(configFilePath)
	if err != nil {
		return nil, fmt.Errorf("cannot read config file %s: %w", configFilePath, err)
	}

	// parse the yaml from the config file
	if err := yaml.Unmarshal(data, &configurationFile); err != nil {
		return nil, fmt.Errorf("cannot parse yaml from config file %s: %w", configFilePath, err)
	}

	return &configurationFile, nil
}
