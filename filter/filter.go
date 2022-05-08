package filter

import "github.com/vikpe/serverstat/qserver"

func Filter[Type any](values []Type, validator func(Type) bool) []Type {
	var result = make([]Type, 0)
	for _, value := range values {
		if validator(value) {
			result = append(result, value)
		}
	}
	return result
}

func GameServers(servers []qserver.GenericServer) []qserver.GenericServer {
	return Filter(servers, isGameServer)
}

func isGameServer(server qserver.GenericServer) bool {
	return server.Version.IsGameServer()
}

func ProxyServers(servers []qserver.GenericServer) []qserver.GenericServer {
	return Filter(servers, isProxyServer)
}

func isProxyServer(server qserver.GenericServer) bool {
	return server.Version.IsProxy()
}

func QtvServers(servers []qserver.GenericServer) []qserver.GenericServer {
	isQtvServer := func(server qserver.GenericServer) bool {
		return server.Version.IsQtv()
	}

	return Filter(servers, isQtvServer)
}
