package transform

import (
	"github.com/vikpe/serverstat/qserver/convert"
	"github.com/vikpe/serverstat/qserver/mvdsv"
	"github.com/vikpe/serverstat/qserver/qtv"
	"github.com/vikpe/serverstat/qserver/qwfwd"
	"metaqtv/geo"
	"metaqtv/provider"
)

type MvdsvWithGeo struct {
	mvdsv.MvdsvExport
	Geo geo.Info
}
type QwfwdWithGeo struct {
	qwfwd.Qwfwd
	Geo geo.Info
}
type QtvWithGeo struct {
	qtv.Qtv
	Geo geo.Info
}

func ToMvdsvServers(serversWithGeo []provider.ServerWithGeo) []MvdsvWithGeo {
	mvdsvServers := make([]MvdsvWithGeo, 0)

	for _, serverWithGeo := range serversWithGeo {
		mvdsvServers = append(mvdsvServers, MvdsvWithGeo{
			MvdsvExport: mvdsv.Export(convert.ToMvdsv(serverWithGeo.GenericServer)),
			Geo:         serverWithGeo.Geo,
		})
	}

	return mvdsvServers
}

func ToQwfwds(serversWithGeo []provider.ServerWithGeo) []QwfwdWithGeo {
	proxies := make([]QwfwdWithGeo, 0)

	for _, serverWithGeo := range serversWithGeo {
		proxies = append(proxies, QwfwdWithGeo{
			Qwfwd: convert.ToQwfwd(serverWithGeo.GenericServer),
			Geo:   serverWithGeo.Geo,
		})
	}

	return proxies
}

func ToQtvServers(serversWithGeo []provider.ServerWithGeo) []QtvWithGeo {
	qtvServers := make([]QtvWithGeo, 0)

	for _, serverWithGeo := range serversWithGeo {
		qtvServers = append(qtvServers, QtvWithGeo{
			Qtv: convert.ToQtv(serverWithGeo.GenericServer),
			Geo: serverWithGeo.Geo,
		})
	}

	return qtvServers
}

func ServerAddressToQtvStreamUrlMap(servers []provider.ServerWithGeo) map[string]string {
	serverToQtv := make(map[string]string, 0)

	for _, server := range servers {
		if "" != server.ExtraInfo.QtvStream.Url {
			serverToQtv[server.Address] = server.ExtraInfo.QtvStream.Url
		}
	}

	return serverToQtv
}

func QtvStreamUrlToServerAddressMap(servers []provider.ServerWithGeo) map[string]string {
	return ReverseStringMap(ServerAddressToQtvStreamUrlMap(servers))
}

func ReverseStringMap(map_ map[string]string) map[string]string {
	reversed := make(map[string]string, 0)
	for key, value := range map_ {
		reversed[value] = key
	}
	return reversed
}
