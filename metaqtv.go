package main

import (
	"log"
	"os"

	"metaqtv/dataprovider"
	"metaqtv/geodb"
	"metaqtv/mhttp"
	"metaqtv/scrape"
)

func main() {
	// config
	conf := getConfig()

	masters, err := getMasterServersFromJsonFile(conf.masterServersFile)

	if err != nil {
		log.Println("Unable to read master_servers.json")
		os.Exit(1)
	}

	// data provider
	scraper := scrape.NewServerScraper(masters)
	scraper.Start()
	geoDatabase, _ := geodb.New()
	provider := dataprovider.New(&scraper, geoDatabase)

	// web server
	webserver := mhttp.NewServer()
	webserver.Endpoints = mhttp.Endpoints{
		"/servers":       mhttp.HandlerBySource(provider.Mvdsv),
		"/qtv":           mhttp.HandlerBySource(provider.Qtv),
		"/qwfwd":         mhttp.HandlerBySource(provider.Qwdwd),
		"/server_to_qtv": mhttp.HandlerBySource(provider.ServerToQtvStream),
		"/qtv_to_server": mhttp.HandlerBySource(provider.QtvStreamToServer),
	}
	webserver.Serve(conf.httpPort)
}
