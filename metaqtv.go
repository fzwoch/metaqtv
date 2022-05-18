package main

import (
	"log"
	"os"

	"metaqtv/geodb"
	"metaqtv/scrape"
	"metaqtv/transform"
	"metaqtv/webserver"
)

func main() {
	// config
	conf := getConfig()

	// read master servers
	masters, err := getMasterServersFromJsonFile(conf.masterServersFile)

	if err != nil {
		log.Println("Unable to read master_servers.json")
		os.Exit(1)
	}

	// geo data
	geoDatabase, _ := geodb.New()

	// server scraper
	serverScraper := scrape.NewServerScraper(masters)
	serverScraper.Start()

	transformer := transform.ServerTransformer{GeoDb: geoDatabase}

	type DataProvider struct {
		scraper     scrape.ServerScraper
		transformer transform.ServerTransformer
		Mvdsv       func() []transform.MvdsvWithGeo
		Qtv         func() []transform.QtvWithGeo
		Qwfwd       func() []transform.QwfwdWithGeo
	}

	provider := DataProvider{
		scraper:     serverScraper,
		transformer: transformer,
	}

	// web api
	webserver.Serve(conf.httpPort, serverScraper)
}
