package main

import (
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	masterstat "github.com/vikpe/qw-masterstat"
	serverstat "github.com/vikpe/qw-serverstat"
)

func main() {
	// conf
	conf := getConfig()

	// get servers
	masters, err := getMasterServersFromJsonFile(conf.masterServersFile)

	if err != nil {
		log.Println("Unable to read master_servers.json")
		os.Exit(1)
	}

	// main
	servers := make([]serverstat.QuakeServer, 0)

	go func() {
		ticker := time.NewTicker(time.Duration(conf.updateInterval) * time.Second)

		for ; true; <-ticker.C {
			wg := sync.WaitGroup{}
			wg.Add(1)

			go func() {
				defer wg.Done()

				serverAddresses := masterstat.StatMany(masters)
				servers = serverstat.StatMany(serverAddresses)
			}()
		}
	}()

	// append geo data to servers
	geoData, err := getGeoData()

	if err != nil {
		log.Println("Unable to download geo data.json")
		os.Exit(1)
	}

	type ServerGeo struct {
		serverstat.QuakeServer
		Geo GeoInfo
	}

	appendGeo := func(servers []serverstat.QuakeServer) []ServerGeo {
		serversWithGeo := make([]ServerGeo, 0)

		for _, server := range servers {
			serverIp := strings.Split(server.Address, ":")[0]
			serversWithGeo = append(serversWithGeo, ServerGeo{
				QuakeServer: server,
				Geo:         geoData[serverIp],
			})
		}

		return serversWithGeo
	}

	// http
	handlerByFilter := func(filterFunc func([]serverstat.QuakeServer) []serverstat.QuakeServer) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			httpJsonResponse(appendGeo(filterFunc(servers)), w, r)
		}
	}

	handlerByMapping := func(mapFunc func([]serverstat.QuakeServer) map[string]string) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) { httpJsonResponse(mapFunc(servers), w, r) }
	}

	api := make(map[string]http.HandlerFunc, 0)
	api["/api/v3/servers"] = handlerByFilter(isNormalServerFilter)
	api["/api/v3/proxies"] = handlerByFilter(isProxyServerFilter)
	api["/api/v3/qtv"] = handlerByFilter(isQtvServerFilter)
	api["/api/v3/server_to_qtv"] = handlerByMapping(serverAddressToQtvMap)
	api["/api/v3/qtv_to_server"] = handlerByMapping(qtvToServerAddressMap)

	cacheClient := getHttpCacheClient()
	for url, handler := range api {
		http.Handle(url, cacheClient.Middleware(handler))
	}

	http.ListenAndServe(":"+strconv.Itoa(conf.httpPort), nil)
}
