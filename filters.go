package main

import "strings"

func isNormalServer(server QuakeServer) bool {
	return !(isQtvServer(server) || isProxyServer(server))
}

func isNormalServerFilter(servers []QuakeServer) []QuakeServer {
	return Filter(servers, isNormalServer)
}

func isProxyServer(server QuakeServer) bool {
	return strings.HasPrefix(server.Settings["*version"], "qwfwd")
}

func isProxyServerFilter(servers []QuakeServer) []QuakeServer {
	return Filter(servers, isProxyServer)
}

func isQtvServer(server QuakeServer) bool {
	return strings.HasPrefix(server.Settings["*version"], "QTV")
}

func isQtvServerFilter(servers []QuakeServer) []QuakeServer {
	return Filter(servers, isQtvServer)
}
