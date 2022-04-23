// Copyright (C) 2019-2022 Florian Zwoch <fzwoch@gmail.com>
//
// see https://github.com/eb/metaqtv
// for jogi - get well soon <3

package main

import (
	"net/http"
	"strconv"
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

	// http
	handlerByFilter := func(validator func(QuakeServer) bool) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) { response(Filter(servers, validator), w, r) }
	}

	handlerByMapTransform := func(mapTransform func([]QuakeServer) map[string]string) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) { response(mapTransform(servers), w, r) }
	}

	api := make(map[string]http.HandlerFunc, 0)
	api["/api/v3/servers"] = handlerByFilter(isNormalServer)
	api["/api/v3/proxies"] = handlerByFilter(isProxyServer)
	api["/api/v3/qtv"] = handlerByFilter(isQtvServer)
	api["/api/v3/server_to_qtv"] = handlerByMapTransform(serverAddressToQtvMap)
	api["/api/v3/qtv_to_server"] = handlerByMapTransform(qtvToServerAddressMap)

	cacheClient := getHttpCacheClient()
	for url, handler := range api {
		http.Handle(url, cacheClient.Middleware(handler))
	}

	http.ListenAndServe(":"+strconv.Itoa(conf.httpPort), nil)
}
