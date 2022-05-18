package transform

import (
	"github.com/vikpe/serverstat/qserver"
	"github.com/vikpe/serverstat/qserver/convert"
	"github.com/vikpe/serverstat/qserver/mvdsv"
	"github.com/vikpe/serverstat/qserver/qtv"
	"github.com/vikpe/serverstat/qserver/qwfwd"
	"metaqtv/geodb"
)

type MvdsvWithGeo struct {
	mvdsv.MvdsvExport
	Geo geodb.Info
}
type QwfwdWithGeo struct {
	qwfwd.QwfwdExport
	Geo geodb.Info
}
type QtvWithGeo struct {
	qtv.QtvExport
	Geo geodb.Info
}

type ServerTransformer struct {
	GeoDb geodb.Database
}

func (st ServerTransformer) ToMvdsvServers(servers []qserver.GenericServer) []MvdsvWithGeo {
	mvdsvServers := make([]MvdsvWithGeo, 0)

	for _, server := range servers {
		if server.Version.IsMvdsv() {
			mvdsvServers = append(mvdsvServers, MvdsvWithGeo{
				MvdsvExport: mvdsv.Export(convert.ToMvdsv(server)),
				Geo:         st.GeoDb.GetByAddress(server.Address),
			})
		}
	}

	return mvdsvServers
}

func (st ServerTransformer) ToQwfwds(servers []qserver.GenericServer) []QwfwdWithGeo {
	proxies := make([]QwfwdWithGeo, 0)

	for _, server := range servers {
		if server.Version.IsQwfwd() {
			proxies = append(proxies, QwfwdWithGeo{
				QwfwdExport: qwfwd.Export(convert.ToQwfwd(server)),
				Geo:         st.GeoDb.GetByAddress(server.Address),
			})
		}
	}

	return proxies
}

func (st ServerTransformer) ToQtvServers(servers []qserver.GenericServer) []QtvWithGeo {
	qtvServers := make([]QtvWithGeo, 0)

	for _, server := range servers {
		qtvServers = append(qtvServers, QtvWithGeo{
			QtvExport: qtv.Export(convert.ToQtv(server)),
			Geo:       st.GeoDb.GetByAddress(server.Address),
		})
	}

	return qtvServers
}

func ServerAddressToQtvStreamUrlMap(servers []qserver.GenericServer) map[string]string {
	serverToQtv := make(map[string]string, 0)

	for _, server := range servers {
		if "" != server.ExtraInfo.QtvStream.Url {
			serverToQtv[server.Address] = server.ExtraInfo.QtvStream.Url
		}
	}

	return serverToQtv
}

func QtvStreamUrlToServerAddressMap(servers []qserver.GenericServer) map[string]string {
	return reverseStringMap(ServerAddressToQtvStreamUrlMap(servers))
}

func reverseStringMap(map_ map[string]string) map[string]string {
	reversed := make(map[string]string, 0)
	for key, value := range map_ {
		reversed[value] = key
	}
	return reversed
}
