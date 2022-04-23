package main

import "flag"

type AppConfig struct {
	httpPort          int
	updateInterval    int
	timeout           int
	retries           int
	masterServersFile string
	keepalive         int
}

func getConfig() AppConfig {
	var (
		httpPort       int
		updateInterval int
		timeout        int
		retries        int
		keepalive      int
	)

	flag.IntVar(&httpPort, "port", 3000, "HTTP listen port")
	flag.IntVar(&updateInterval, "interval", 60, "Update interval in seconds")
	flag.IntVar(&timeout, "timeout", 500, "UDP timeout in milliseconds")
	flag.IntVar(&retries, "retry", 5, "UDP retry count")
	flag.IntVar(&keepalive, "keepalive", 3, "Keep server alive for N tries")
	flag.Parse()

	return AppConfig{
		httpPort:          httpPort,
		updateInterval:    updateInterval,
		timeout:           timeout,
		retries:           retries,
		keepalive:         keepalive,
		masterServersFile: "master_servers.json",
	}
}
