package main

import (
	"log"
	"os"

	"metaqtv/geo"
	"metaqtv/provider"
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
	geoDatabase, _ := geo.NewDatabase()

	// server scraper
	serverScraper := provider.New(masters, geoDatabase)
	serverScraper.Start()

	// web api
	webserver.Serve(conf.httpPort, serverScraper.Servers)
}
