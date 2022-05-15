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
	return Filter(servers, isMvdsvServer)
}

func isMvdsvServer(server geo.ServerWithGeo) bool {
	return server.Version.IsMvdsv()
}

func Qwforwards(servers []geo.ServerWithGeo) []geo.ServerWithGeo {
	return Filter(servers, isQwfwd)
}

func isQwfwd(server geo.ServerWithGeo) bool {
	return server.Version.IsQwfwd()
}

func QtvServers(servers []geo.ServerWithGeo) []geo.ServerWithGeo {
	return Filter(servers, isQtvServer)
}

func isQtvServer(server geo.ServerWithGeo) bool {
	return server.Version.IsQtv()
}
