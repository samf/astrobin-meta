package main

import (
	"errors"
	"fmt"
	"io/fs"
	"os"

	"gopkg.in/yaml.v3"
)

// config is the on-disk YAML configuration. For now it holds only the
// filter-name -> AstroBin-filter-ID map, but a struct leaves room to grow.
type config struct {
	Filters map[string]int `yaml:"filters"`
}

// loadFilterMap reads the YAML config at path and returns the filter-name ->
// AstroBin-filter-ID map.
func loadFilterMap(path string) (map[string]int, error) {
	data, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, fmt.Errorf(`filter config file not found: %s

Create one (YAML) mapping filter names -> AstroBin filter IDs, e.g.:

filters:
  L: 33995
  Ha: 43627
  OIII: 43628`, path)
	}
	if err != nil {
		return nil, fmt.Errorf("reading config %s: %w", path, err)
	}

	var cfg config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config %s: %w", path, err)
	}

	if len(cfg.Filters) == 0 {
		return nil, fmt.Errorf("config %s has no filters defined under the 'filters:' key", path)
	}

	return cfg.Filters, nil
}
