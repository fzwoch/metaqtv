package main

import "github.com/vikpe/serverstat/qserver"

func filter[Type any](values []Type, validator func(Type) bool) []Type {
	var result = make([]Type, 0)
	for _, value := range values {
		if validator(value) {
			result = append(result, value)
		}
	}
	return result
}

func isGameServerFilter(servers []qserver.GenericServer) []qserver.GenericServer {
	isGameServer := func(server qserver.GenericServer) bool {
		return server.Version.IsGameServer()
	}

	return filter(servers, isGameServer)
}

func isProxyServerFilter(servers []qserver.GenericServer) []qserver.GenericServer {
	isProxyServer := func(server qserver.GenericServer) bool {
		return server.Version.IsProxy()
	}

	return filter(servers, isProxyServer)
}

func isQtvServerFilter(servers []qserver.GenericServer) []qserver.GenericServer {
	isQtvServer := func(server qserver.GenericServer) bool {
		return server.Version.IsQtv()
	}

	return filter(servers, isQtvServer)
}
