package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/vikpe/masterstat"
	"github.com/vikpe/serverstat"
	"metaqtv/geo"
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
	geoDatabase, err := geo.NewDatabase()

	// main loop
	serversWithGeo := make([]geo.ServerWithGeo, 0)

	go func() {
		ticker := time.NewTicker(time.Duration(conf.updateInterval) * time.Second)

		for ; true; <-ticker.C {
			go func() {
				serverAddresses, err := masterstat.GetServerAddressesFromMany(masters)

				if err != nil {
					log.Println("ERROR:", err)
					return
				}
				servers := serverstat.GetInfoFromMany(serverAddresses)
				serversWithGeo = geo.AppendGeo(servers, geoDatabase)
			}()
		}
	}()

	// web api
	serverAddr := fmt.Sprintf(":%d", conf.httpPort)
	webserver.Serve(serverAddr, &serversWithGeo)
}
