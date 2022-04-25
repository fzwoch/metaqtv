package main

import (
	serverstat "github.com/vikpe/qw-serverstat"
)

func serverAddressToQtvMap(servers []serverstat.QuakeServer) map[string]string {
	normalServers := filter(servers, isNormalServer)
	serverToQtv := make(map[string]string, 0)

	for _, server := range normalServers {
		if server.QtvAddress != "" {
			serverToQtv[server.Address] = server.QtvAddress
		}
	}

	return serverToQtv
}

func qtvToServerAddressMap(servers []serverstat.QuakeServer) map[string]string {
	return reverseStringMap(serverAddressToQtvMap(servers))
}

func reverseStringMap(map_ map[string]string) map[string]string {
	reversed := make(map[string]string, 0)
	for key, value := range map_ {
		reversed[value] = key
	}
	return reversed
}
