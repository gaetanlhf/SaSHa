package main

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
)

func loadConfig(path string) (Config, error) {
	var config Config

	data, err := os.ReadFile(path)
	if err != nil {
		return config, err
	}

	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return config, err
	}

	applyGlobalSettings(&config)

	importErr := ProcessImports(&config, path)

	if importErr != nil && len(config.Groups) == 0 && len(config.Hosts) == 0 {
		return config, fmt.Errorf("failed to process imports: %w", importErr)
	}

	return config, importErr
}
