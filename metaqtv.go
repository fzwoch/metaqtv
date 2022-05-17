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
	serverAddresses := make([]string, 0)
	masterUpdateInterval := 600

	go func() {
		ticker := time.NewTicker(time.Duration(1) * time.Second)
		tick := -1

		for ; true; <-ticker.C {
			tick++

			go func() {
				currentTick := tick

				isTimeToUpdateFromMasters := 0 == currentTick

				if isTimeToUpdateFromMasters {
					serverAddresses, err = masterstat.GetServerAddressesFromMany(masters)

					if err != nil {
						log.Println("ERROR:", err)
						return
					}
				}

				isTimeToUpdateServers := currentTick%conf.updateInterval == 0

				if isTimeToUpdateServers {
					servers := serverstat.GetInfoFromMany(serverAddresses)
					serversWithGeo = geo.AppendGeo(servers, geoDatabase)
				}
			}()

			if tick == masterUpdateInterval {
				tick = 0
			}
		}
	}()

	// web api
	serverAddr := fmt.Sprintf(":%d", conf.httpPort)
	webserver.Serve(serverAddr, &serversWithGeo)
}
