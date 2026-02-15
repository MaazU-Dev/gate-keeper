package config

import (
	"encoding/json"
	"fmt"
	"os"
)

func LoadServices(path string) ([]Service, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var services []Service
	if err := json.NewDecoder(f).Decode(&services); err != nil {
		return nil, err
	}

	if len(services) == 0 {
		return nil, fmt.Errorf("no services found in config")
	}

	return services, nil
}
