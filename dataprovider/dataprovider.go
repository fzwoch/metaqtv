package dataprovider

import (
	"github.com/vikpe/serverstat/qserver/convert"
	"github.com/vikpe/serverstat/qserver/mvdsv"
	"github.com/vikpe/serverstat/qserver/qtv"
	"github.com/vikpe/serverstat/qserver/qwfwd"
	"metaqtv/geodb"
	"metaqtv/scrape"
)

type DataProvider struct {
	scraper *scrape.ServerScraper
	geoDb   geodb.Database
}

func New(scraper *scrape.ServerScraper, geoDb geodb.Database) DataProvider {
	return DataProvider{
		scraper: scraper,
		geoDb:   geoDb,
	}
}

func (dp DataProvider) Mvdsv() any {
	result := make([]mvdsv.MvdsvExport, 0)

	for _, server := range dp.scraper.Servers() {
		if server.Version.IsMvdsv() {
			mvdsvExport := convert.ToMvdsvExport(server)
			mvdsvExport.Geo = dp.geoDb.GetByAddress(server.Address)
			result = append(result, mvdsvExport)
		}
	}

	return result
}

func (dp DataProvider) Qtv() any {
	result := make([]qtv.QtvExport, 0)

	for _, server := range dp.scraper.Servers() {
		if server.Version.IsQtv() {
			qtvExport := convert.ToQtvExport(server)
			qtvExport.Geo = dp.geoDb.GetByAddress(server.Address)
			result = append(result, qtvExport)
		}
	}

	return result
}

func (dp DataProvider) Qwdwd() any {
	result := make([]qwfwd.QwfwdExport, 0)

	for _, server := range dp.scraper.Servers() {
		if server.Version.IsQtv() {
			qwfwdExport := convert.ToQwfwdExport(server)
			qwfwdExport.Geo = dp.geoDb.GetByAddress(server.Address)
			result = append(result, qwfwdExport)
		}
	}

	return result
}

func (dp DataProvider) ServerToQtvStream() any {
	serverToQtv := make(map[string]string, 0)

	for _, server := range dp.scraper.Servers() {
		if "" != server.ExtraInfo.QtvStream.Url {
			serverToQtv[server.Address] = server.ExtraInfo.QtvStream.Url
		}
	}

	return serverToQtv
}

func (dp DataProvider) QtvStreamToServer() any {
	qtvToServer := make(map[string]string, 0)

	for _, server := range dp.scraper.Servers() {
		if "" != server.ExtraInfo.QtvStream.Url {
			qtvToServer[server.ExtraInfo.QtvStream.Url] = server.Address
		}
	}

	return qtvToServer
}
