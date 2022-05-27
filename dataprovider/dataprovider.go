package dataprovider

import (
	"github.com/vikpe/serverstat/qserver"
	"github.com/vikpe/serverstat/qserver/convert"
	"github.com/vikpe/serverstat/qserver/mvdsv"
	"metaqtv/scrape"
)

type DataProvider struct {
	scraper *scrape.ServerScraper
}

func New(scraper *scrape.ServerScraper) DataProvider {
	return DataProvider{
		scraper: scraper,
	}
}

func (dp DataProvider) Generic() []qserver.GenericServer {
	return dp.scraper.Servers()
}

func (dp DataProvider) Mvdsv() []mvdsv.MvdsvExport {
	result := make([]mvdsv.MvdsvExport, 0)

	for _, server := range dp.scraper.Servers() {
		if server.Version.IsMvdsv() {
			mvdsvExport := convert.ToMvdsvExport(server)

			if mvdsvExport.QtvStream.Address != "" {
				result = append(result, mvdsvExport)
			}
		}
	}

	return result
}
