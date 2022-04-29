package main

import (
	"strings"

	"github.com/vikpe/qw-serverstat/quakeserver"
)

func filter[Type any](values []Type, validator func(Type) bool) []Type {
	var result = make([]Type, 0)
	for _, value := range values {
		if validator(value) {
			result = append(result, value)
		}
	}
	return result
}

func isNormalServer(server quakeserver.QuakeServer) bool {
	return !(isQtvServer(server) || isProxyServer(server))
}

func isNormalServerFilter(servers []quakeserver.QuakeServer) []quakeserver.QuakeServer {
	return filter(servers, isNormalServer)
}

func isProxyServer(server quakeserver.QuakeServer) bool {
	return strings.HasPrefix(server.Settings["*version"], "qwfwd")
}

func isProxyServerFilter(servers []quakeserver.QuakeServer) []quakeserver.QuakeServer {
	return filter(servers, isProxyServer)
}

func isQtvServer(server quakeserver.QuakeServer) bool {
	return strings.HasPrefix(server.Settings["*version"], "QTV")
}

func isQtvServerFilter(servers []quakeserver.QuakeServer) []quakeserver.QuakeServer {
	return filter(servers, isQtvServer)
}
