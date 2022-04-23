package main

func serverAddressToQtvMap(servers []QuakeServer) map[string]string {
	normalServers := Filter(servers, isNormalServer)
	serverToQtv := make(map[string]string, 0)

	for _, server := range normalServers {
		if server.QtvAddress != "" {
			serverToQtv[server.Address] = server.QtvAddress
		}
	}

	return serverToQtv
}

func qtvToServerAddressMap(servers []QuakeServer) map[string]string {
	return reverseStringMap(serverAddressToQtvMap(servers))
}
