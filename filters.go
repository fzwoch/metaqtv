package main

import (
	"strings"

	serverstat "github.com/vikpe/qw-serverstat"
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

func isNormalServer(server serverstat.QuakeServer) bool {
	return !(isQtvServer(server) || isProxyServer(server))
}

func isNormalServerFilter(servers []serverstat.QuakeServer) []serverstat.QuakeServer {
	return filter(servers, isNormalServer)
}

func isProxyServer(server serverstat.QuakeServer) bool {
	return strings.HasPrefix(server.Settings["*version"], "qwfwd")
}

func isProxyServerFilter(servers []serverstat.QuakeServer) []serverstat.QuakeServer {
	return filter(servers, isProxyServer)
}

func isQtvServer(server serverstat.QuakeServer) bool {
	return strings.HasPrefix(server.Settings["*version"], "QTV")
}

func isQtvServerFilter(servers []serverstat.QuakeServer) []serverstat.QuakeServer {
	return filter(servers, isQtvServer)
}
