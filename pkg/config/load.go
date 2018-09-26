package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/Scalify/website-content-watcher/pkg/api"
	"github.com/ghodss/yaml"
)

// Load a config file from disk
func Load(file string) (*api.Config, error) {
	ext := strings.ToLower(filepath.Ext(file))
	if ext == ".yaml" || ext == ".yml" {
		return loadYAML(file)
	}

	if ext == ".json" {
		return loadJSON(file)
	}

	return nil, fmt.Errorf("cannot handle file extension %s", ext)
}

func loadYAML(file string) (*api.Config, error) {
	yamlBytes, err := read(file)
	if err != nil {
		return nil, err
	}

	jsonBytes, err := yaml.YAMLToJSON(yamlBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to convert json to yaml: %v", err)
	}

	return parse(jsonBytes)

}
func loadJSON(file string) (*api.Config, error) {
	b, err := read(file)
	if err != nil {
		return nil, err
	}

	return parse(b)
}

func read(file string) ([]byte, error) {
	b, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("failed to load file %s: %v", file, err)
	}

	return b, nil
}

func parse(b []byte) (*api.Config, error) {
	cfg := &api.Config{}

	if err := json.Unmarshal(b, cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %v", err)
	}

	return cfg, nil
}
