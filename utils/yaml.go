package utils

import (
	"os"

	"gopkg.in/yaml.v3"
)

func ReadYAML(path string, data interface{}) error {

	file, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	return yaml.Unmarshal(file, data)
}

func WriteYAML(path string, data interface{}) error {
	file, err := yaml.Marshal(data)

	if err != nil {
		return err
	}

	return os.WriteFile(path, file, 0644)
}
