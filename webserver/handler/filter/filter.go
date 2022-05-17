package filter

import (
	"metaqtv/provider"
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

func MvdsvServers(servers []provider.ServerWithGeo) []provider.ServerWithGeo {
	return Filter(servers, isMvdsvServer)
}

func isMvdsvServer(server provider.ServerWithGeo) bool {
	return server.Version.IsMvdsv()
}

func Qwforwards(servers []provider.ServerWithGeo) []provider.ServerWithGeo {
	return Filter(servers, isQwfwd)
}

func isQwfwd(server provider.ServerWithGeo) bool {
	return server.Version.IsQwfwd()
}

func QtvServers(servers []provider.ServerWithGeo) []provider.ServerWithGeo {
	return Filter(servers, isQtvServer)
}

func isQtvServer(server provider.ServerWithGeo) bool {
	return server.Version.IsQtv()
}

func ForstressOneServers(servers []provider.ServerWithGeo) []provider.ServerWithGeo {
	return Filter(servers, isForstressOneServer)
}

func isForstressOneServer(server provider.ServerWithGeo) bool {
	return server.Version.IsFortressOne()
}
