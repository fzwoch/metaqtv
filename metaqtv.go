package main

import (
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/vikpe/masterstat"
	"github.com/vikpe/serverstat"
	"github.com/vikpe/serverstat/qserver"
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

	// main loop
	servers := make([]qserver.GenericServer, 0)

	go func() {
		ticker := time.NewTicker(time.Duration(conf.updateInterval) * time.Second)

		for ; true; <-ticker.C {
			wg := sync.WaitGroup{}
			wg.Add(1)

			go func() {
				defer wg.Done()

				serverAddresses, err := masterstat.GetServerAddressesFromMany(masters)

				if err != nil {
					log.Println("ERROR:", err)
					return
				}
				servers = serverstat.GetInfoFromMany(serverAddresses)
			}()
		}
	}()

	// web api
	serverAddr := fmt.Sprintf(":%d", conf.httpPort)
	webserver.Serve(serverAddr, &servers)
}
