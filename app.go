package main

import "flag"

type AppConfig struct {
	httpPort          int
	updateInterval    int
	timeout           int
	retries           int
	masterServersFile string
}

func getConfig() AppConfig {
	var (
		httpPort       int
		updateInterval int
		timeout        int
		retries        int
	)

	flag.IntVar(&httpPort, "port", 3000, "HTTP listen port")
	flag.IntVar(&updateInterval, "interval", 60, "Update interval in seconds")
	flag.IntVar(&timeout, "timeout", 500, "Timeout in milliseconds")
	flag.IntVar(&retries, "retry", 3, "Retry count")
	flag.Parse()

	return AppConfig{
		httpPort:          httpPort,
		updateInterval:    updateInterval,
		timeout:           timeout,
		retries:           retries,
		masterServersFile: "master_servers.json",
	}
}
