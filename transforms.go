package main

import "github.com/vikpe/serverstat/qserver"

func serverAddressToQtvMap(servers []qserver.GenericServer) map[string]string {
	gameServers := isGameServerFilter(servers)
	serverToQtv := make(map[string]string, 0)

	for _, server := range gameServers {
		if "" != server.ExtraInfo.QtvStream.Url {
			serverToQtv[server.Address] = server.ExtraInfo.QtvStream.Url
		}
	}

	return serverToQtv
}

func qtvToServerAddressMap(servers []qserver.GenericServer) map[string]string {
	return reverseStringMap(serverAddressToQtvMap(servers))
}

func reverseStringMap(map_ map[string]string) map[string]string {
	reversed := make(map[string]string, 0)
	for key, value := range map_ {
		reversed[value] = key
	}
	return reversed
}
