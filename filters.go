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
	return filter(servers, func(server qserver.GenericServer) bool {
		return server.Version.IsGameServer()
	})
}

func isProxyServerFilter(servers []qserver.GenericServer) []qserver.GenericServer {
	return filter(servers, func(server qserver.GenericServer) bool {
		return server.Version.IsProxy()
	})
}

func isQtvServerFilter(servers []qserver.GenericServer) []qserver.GenericServer {
	return filter(servers, func(server qserver.GenericServer) bool {
		return server.Version.IsQtv()
	})
}
