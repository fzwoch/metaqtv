package main

import (
	"encoding/json"
	"flag"
	"os"
)

type AppConfig struct {
	httpPort          int
	updateInterval    int
	masterServersFile string
}

func getConfig() AppConfig {
	var (
		httpPort       int
		updateInterval int
	)

	flag.IntVar(&httpPort, "port", 3000, "HTTP listen port")
	flag.IntVar(&updateInterval, "interval", 60, "Update interval in seconds")
	flag.Parse()

	return AppConfig{
		httpPort:          httpPort,
		updateInterval:    updateInterval,
		masterServersFile: "master_servers.json",
	}
}

func getMasterServersFromJsonFile(filePath string) ([]string, error) {
	result := make([]string, 0)

	jsonFile, err := os.ReadFile(filePath)
	if err != nil {
		return result, err
	}

	err = json.Unmarshal(jsonFile, &result)
	if err != nil {
		return result, err
	}

	return result, nil
}
