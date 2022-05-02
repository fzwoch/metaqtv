package main

import "github.com/vikpe/qw-serverstat/qserver"

func filter[Type any](values []Type, validator func(Type) bool) []Type {
	var result = make([]Type, 0)
	for _, value := range values {
		if validator(value) {
			result = append(result, value)
		}
	}
	return result
}

func isNormalServerFilter(servers []qserver.GenericServer) []qserver.GenericServer {
	return filter(servers, qserver.IsGameServer)
}

func isProxyServerFilter(servers []qserver.GenericServer) []qserver.GenericServer {
	return filter(servers, qserver.IsProxyServer)
}

func isQtvServerFilter(servers []qserver.GenericServer) []qserver.GenericServer {
	return filter(servers, qserver.IsQtvServer)
}
