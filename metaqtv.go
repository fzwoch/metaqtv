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
	serversHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response(Filter(servers, isNormalServer), w, r)
	})

	proxiesHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response(Filter(servers, isProxyServer), w, r)
	})

	qtvHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response(Filter(servers, isQtvServer), w, r)
	})

	cacheClient := getHttpCacheClient()
	http.Handle("/api/v3/servers", cacheClient.Middleware(serversHandler))
	http.Handle("/api/v3/proxies", cacheClient.Middleware(proxiesHandler))
	http.Handle("/api/v3/qtv", cacheClient.Middleware(qtvHandler))
	http.ListenAndServe(":"+strconv.Itoa(conf.httpPort), nil)
}
