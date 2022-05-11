package filter

import (
	"metaqtv/geo"
)

func Filter[Type any](values []Type, validator func(Type) bool) []Type {
	var result = make([]Type, 0)
	for _, value := range values {
		if validator(value) {
			result = append(result, value)
		}
	}
	return result
}

func MvdsvServers(servers []geo.ServerWithGeo) []geo.ServerWithGeo {
	return Filter(servers, isGameServer)
}

func isGameServer(server geo.ServerWithGeo) bool {
	return server.Version.IsGameServer()
}

func ProxyServers(servers []geo.ServerWithGeo) []geo.ServerWithGeo {
	return Filter(servers, isProxyServer)
}

func isProxyServer(server geo.ServerWithGeo) bool {
	return server.Version.IsProxy()
}

func QtvServers(servers []geo.ServerWithGeo) []geo.ServerWithGeo {
	isQtvServer := func(server geo.ServerWithGeo) bool {
		return server.Version.IsQtv()
	}

	return Filter(servers, isQtvServer)
}
