package provider

import (
	"log"
	"strings"
	"time"

	"github.com/vikpe/masterstat"
	"github.com/vikpe/serverstat"
	"github.com/vikpe/serverstat/qserver"
	"metaqtv/geo"
)

type ServerWithGeo struct {
	qserver.GenericServer
	Geo geo.Info
}

func AppendGeo(servers []qserver.GenericServer, geoDb geo.Database) []ServerWithGeo {
	serversWithGeo := make([]ServerWithGeo, 0)

	for _, server := range servers {
		ip := strings.Split(server.Address, ":")[0]
		serversWithGeo = append(serversWithGeo, ServerWithGeo{
			GenericServer: server,
			Geo:           geoDb.Get(ip),
		})
	}

	return serversWithGeo
}

type ServerScraper struct {
	masters    []string
	servers    []ServerWithGeo
	shouldStop bool
	geoDb      geo.Database
}

func New(masters []string, geoDb geo.Database) ServerScraper {
	return ServerScraper{
		masters:    masters,
		servers:    make([]ServerWithGeo, 0),
		shouldStop: false,
		geoDb:      geoDb,
	}
}

func (sp *ServerScraper) Servers() []ServerWithGeo {
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
					servers := serverstat.GetInfoFromMany(serverAddresses)
					sp.servers = AppendGeo(servers, sp.geoDb)
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
