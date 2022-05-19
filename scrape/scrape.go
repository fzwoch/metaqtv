package scrape

import (
	"log"
	"time"

	"github.com/vikpe/masterstat"
	"github.com/vikpe/serverstat"
	"github.com/vikpe/serverstat/qserver"
)

type ServerIndex map[string]qserver.GenericServer

func NewServerIndex(servers []qserver.GenericServer) ServerIndex {
	index := make(ServerIndex, 0)

	for _, server := range servers {
		index[server.Address] = server
	}

	return index
}

func (index ServerIndex) Servers() []qserver.GenericServer {
	servers := make([]qserver.GenericServer, 0)

	for _, server := range index {
		servers = append(servers, server)
	}

	return servers
}

func (index ServerIndex) ActiveAddresses() []string {
	activeAddresses := make([]string, 0)

	for _, server := range index.Servers() {
		if hasPlayers(server) {
			activeAddresses = append(activeAddresses, server.Address)
		}
	}

	return activeAddresses
}

func hasPlayers(server qserver.GenericServer) bool {
	for _, c := range server.Clients {
		if !c.IsSpectator() && !c.IsBot() {
			return true
		}
	}

	return false
}

func (index ServerIndex) Update(servers []qserver.GenericServer) {
	for _, server := range servers {
		index[server.Address] = server
	}
}

type ServerScraper struct {
	masters         []string
	index           ServerIndex
	serverAddresses []string
	shouldStop      bool
}

func NewServerScraper(masters []string) ServerScraper {
	return ServerScraper{
		masters:         masters,
		index:           make(ServerIndex, 0),
		serverAddresses: make([]string, 0),
		shouldStop:      false,
	}
}

func (sp *ServerScraper) Servers() []qserver.GenericServer {
	return sp.index.Servers()
}

func (sp *ServerScraper) Start() {
	masterUpdateInterval := 600
	serverUpdateInterval := 30
	activeServerUpdateInterval := 5

	serverAddresses := make([]string, 0)
	sp.shouldStop = false

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

				isTimeToUpdateAllServers := currentTick%serverUpdateInterval == 0
				isTimeToUpdateActiveServers := currentTick%activeServerUpdateInterval == 0

				if isTimeToUpdateAllServers {
					sp.index = NewServerIndex(serverstat.GetInfoFromMany(serverAddresses))
				} else if isTimeToUpdateActiveServers {
					activeAddresses := sp.index.ActiveAddresses()
					sp.index.Update(serverstat.GetInfoFromMany(activeAddresses))
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
