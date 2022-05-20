package main

import (
	"fmt"
	"log"
	"os"

	apiVersion1 "metaqtv/api/v1"
	apiVersion2 "metaqtv/api/v2"
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
	dataProvider := dataprovider.New(&scraper, geoDatabase)

	// api versions
	apiVersions := []mhttp.Api{
		apiVersion1.New("v1", &dataProvider),
		apiVersion2.New("v2", &dataProvider),
	}

	// merge endpoints
	endpoints := make(mhttp.Endpoints, 0)

	for _, api := range apiVersions {
		for url, handler := range api.Endpoints {
			fullUrl := fmt.Sprintf("/%s/%s", api.BaseUrl, url)
			endpoints[fullUrl] = handler
		}
	}

	mhttp.Serve(conf.httpPort, endpoints)
}
