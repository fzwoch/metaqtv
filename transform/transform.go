package transform

import (
	"github.com/vikpe/serverstat/qserver"
	"metaqtv/filter"
)

func ServerAddressToQtvMap(servers []qserver.GenericServer) map[string]string {
	gameServers := filter.GameServers(servers)
	serverToQtv := make(map[string]string, 0)

	for _, server := range gameServers {
		if "" != server.ExtraInfo.QtvStream.Url {
			serverToQtv[server.Address] = server.ExtraInfo.QtvStream.Url
		}
	}

	return serverToQtv
}

func QtvToServerAddressMap(servers []qserver.GenericServer) map[string]string {
	return reverseStringMap(ServerAddressToQtvMap(servers))
}

func reverseStringMap(map_ map[string]string) map[string]string {
	reversed := make(map[string]string, 0)
	for key, value := range map_ {
		reversed[value] = key
	}
	return reversed
}
