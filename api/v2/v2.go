package v2

import (
	"net/http"

	"github.com/vikpe/serverstat/qserver"
	"github.com/vikpe/serverstat/qserver/mvdsv"
	"github.com/vikpe/serverstat/qserver/qtv"
	"github.com/vikpe/serverstat/qserver/qwfwd"
	"metaqtv/dataprovider"
	"metaqtv/mhttp"
)

func MvdsvHandler(serverSource func() []mvdsv.MvdsvExport) http.HandlerFunc {
	return mhttp.CreateHandler(func() any { return serverSource() })
}

func QtvHandler(serverSource func() []qtv.QtvExport) http.HandlerFunc {
	return mhttp.CreateHandler(func() any { return serverSource() })
}

func QwfwdHandler(serverSource func() []qwfwd.QwfwdExport) http.HandlerFunc {
	return mhttp.CreateHandler(func() any { return serverSource() })
}

func ServerToQtvHandler(serverSource func() []qserver.GenericServer) http.HandlerFunc {
	getServerToQtvMap := func() any {
		serverToQtv := make(map[string]string, 0)
		for _, server := range serverSource() {
			if "" != server.ExtraInfo.QtvStream.Address {
				serverToQtv[server.Address] = server.ExtraInfo.QtvStream.Url()
			}
		}
		return serverToQtv
	}

	return mhttp.CreateHandler(func() any { return getServerToQtvMap() })
}

func QtvToServerHandler(serverSource func() []qserver.GenericServer) http.HandlerFunc {
	getServerToQtvMap := func() any {
		serverToQtv := make(map[string]string, 0)
		for _, server := range serverSource() {
			if "" != server.ExtraInfo.QtvStream.Address {
				serverToQtv[server.ExtraInfo.QtvStream.Url()] = server.Address
			}
		}
		return serverToQtv
	}

	return mhttp.CreateHandler(func() any { return getServerToQtvMap() })
}

func New(baseUrl string, provider *dataprovider.DataProvider) mhttp.Api {
	return mhttp.Api{
		Provider: provider,
		BaseUrl:  baseUrl,
		Endpoints: mhttp.Endpoints{
			"servers":       MvdsvHandler(provider.Mvdsv),
			"qtv":           QtvHandler(provider.Qtv),
			"qwfwd":         QwfwdHandler(provider.Qwdwd),
			"server_to_qtv": ServerToQtvHandler(provider.Generic),
			"qtv_to_server": QtvToServerHandler(provider.Generic),
		},
	}
}
