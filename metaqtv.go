// Copyright (C) 2019-2022 Florian Zwoch <fzwoch@gmail.com>
//
// see https://github.com/eb/metaqtv
// for jogi - get well soon <3

package main

import (
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

func main() {
	// conf
	conf := getConfig()

	// servers
	masters := getMasterServersFromJsonFile(conf.masterServersFile)
	servers := make([]QuakeServer, 0)

	go func() {
		ticker := time.NewTicker(time.Duration(conf.updateInterval) * time.Second)

		for ; true; <-ticker.C {
			wg := sync.WaitGroup{}
			wg.Add(1)

			go func() {
				defer wg.Done()

				serverAddresses := ReadMasterServers(masters, conf.retries, conf.timeout)
				servers = ReadServers(serverAddresses, conf.retries, conf.timeout)

			}()
		}
	}()

	// geo
	geoData := getGeoData()

	appendGeoData := func(servers []QuakeServer) []QuakeServer {
		for index, s := range servers {
			AddressParts := strings.Split(s.Address, ":")
			servers[index].Geo = geoData[AddressParts[0]]
		}
		return servers
	}

	// http
	handlerByFilter := func(filterFunc func([]QuakeServer) []QuakeServer) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			httpJsonResponse(appendGeoData(filterFunc(servers)), w, r)
		}
	}

	handlerByMapping := func(mapFunc func([]QuakeServer) map[string]string) http.HandlerFunc {
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
