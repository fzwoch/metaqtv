package main

import "strings"

func isNormalServer(server QuakeServer) bool {
	return !(isQtvServer(server) || isProxyServer(server))
}

func isProxyServer(server QuakeServer) bool {
	return strings.HasPrefix(server.Settings["*version"], "qwfwd")
}

func isQtvServer(server QuakeServer) bool {
	return strings.HasPrefix(server.Settings["*version"], "QTV")
}
