package main

import (
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/vikpe/masterstat"
	"github.com/vikpe/serverstat"
	"github.com/vikpe/serverstat/qserver"
)

func main() {
	// conf
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
				servers = serverstat.GetServerInfoFromMany(serverAddresses)
			}()
		}
	}()

	// append geo data
	geoData, err := getGeoData()

	if err != nil {
		log.Println("Unable to download geo data.json")
		os.Exit(1)
	}

	type ServerGeo struct {
		qserver.GenericServer
		Geo GeoInfo
	}

	appendGeo := func(servers []qserver.GenericServer) []ServerGeo {
		serversWithGeo := make([]ServerGeo, 0)

		for _, server := range servers {
			serverIp := strings.Split(server.Address, ":")[0]
			serversWithGeo = append(serversWithGeo, ServerGeo{
				GenericServer: server,
				Geo:           geoData[serverIp],
			})
		}

		return serversWithGeo
	}

	// http
	handlerByFilter := func(filterFunc func([]qserver.GenericServer) []qserver.GenericServer) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			httpJsonResponse(appendGeo(filterFunc(servers)), w, r)
		}
	}

	handlerByMapping := func(mapFunc func([]qserver.GenericServer) map[string]string) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) { httpJsonResponse(mapFunc(servers), w, r) }
	}

	api := make(map[string]http.HandlerFunc, 0)
	api["/servers"] = handlerByFilter(isGameServerFilter)
	api["/proxies"] = handlerByFilter(isProxyServerFilter)
	api["/qtv"] = handlerByFilter(isQtvServerFilter)
	api["/server_to_qtv"] = handlerByMapping(serverAddressToQtvMap)
	api["/qtv_to_server"] = handlerByMapping(qtvToServerAddressMap)

	cacheClient := getHttpCacheClient()
	for url, handler := range api {
		http.Handle(url, cacheClient.Middleware(handler))
	}

	http.ListenAndServe(":"+strconv.Itoa(conf.httpPort), nil)
}
