package scrape

import (
	"log"
	"time"

	"github.com/vikpe/masterstat"
	"github.com/vikpe/serverstat"
	"github.com/vikpe/serverstat/qserver"
)

type ServerScraper struct {
	masters    []string
	servers    []qserver.GenericServer
	shouldStop bool
}

func NewServerScraper(masters []string) ServerScraper {
	return ServerScraper{
		masters:    masters,
		servers:    make([]qserver.GenericServer, 0),
		shouldStop: false,
	}
}

func (sp *ServerScraper) Servers() []qserver.GenericServer {
	return sp.servers
}

func (sp *ServerScraper) Start() {
	masterUpdateInterval := 600
	sp.shouldStop = false

	serverAddresses := make([]string, 0)

	go func() {
		ticker := time.NewTicker(time.Duration(1) * time.Second)
		tick := -1

		for ; true; <-ticker.C {
			if sp.shouldStop {
				return
			}

			tick++

			go func() {
				currentTick := tick

				isTimeToUpdateFromMasters := 0 == currentTick

				if isTimeToUpdateFromMasters {
					var err error
					serverAddresses, err = masterstat.GetServerAddressesFromMany(sp.masters)

					if err != nil {
						log.Println("ERROR:", err)
						return
					}
				}

				isTimeToUpdateServers := currentTick%10 == 0

				if isTimeToUpdateServers {
					sp.servers = serverstat.GetInfoFromMany(serverAddresses)
				}
			}()

			if tick == masterUpdateInterval {
				tick = 0
			}
		}
	}()
}

func (sp *ServerScraper) Stop() {
	sp.shouldStop = true
}
