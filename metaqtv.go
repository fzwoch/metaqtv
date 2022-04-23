// Copyright (C) 2019-2022 Florian Zwoch <fzwoch@gmail.com>
//
// see https://github.com/eb/metaqtv
// for jogi - get well soon <3

package main

import (
	"encoding/json"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/victorspringer/http-cache"
	"github.com/victorspringer/http-cache/adapter/memory"
)

func getMasterServersFromJsonFile(filePath string) []SocketAddress {
	jsonFile, err := os.ReadFile(filePath)
	panicIf(err)

	var result []SocketAddress
	err = json.Unmarshal(jsonFile, &result)
	panicIf(err)

	return result
}

const bufferMaxSize = 8192

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
		respond(servers, w, r)
	})

	proxiesHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		respond(Filter(servers, isProxy), w, r)
	})

	qtvHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		respond(Filter(servers, isQtv), w, r)
	})

	cacheClient := getHttpCacheClient()
	http.Handle("/api/v3/servers", cacheClient.Middleware(serversHandler))
	http.Handle("/api/v3/proxies", cacheClient.Middleware(proxiesHandler))
	http.Handle("/api/v3/qtv", cacheClient.Middleware(qtvHandler))
	http.ListenAndServe(":"+strconv.Itoa(conf.httpPort), nil)
}

func isProxy(server QuakeServer) bool {
	return strings.HasPrefix(server.Settings["*version"], "qwfwd")
}

func isQtv(server QuakeServer) bool {
	return strings.HasPrefix(server.Settings["*version"], "QTV")
}

func respond(servers []QuakeServer, response http.ResponseWriter, request *http.Request) {
	response.Header().Set("Content-Type", "application/json")

	serversAsJson, _ := json.MarshalIndent(servers, "", "\t")
	responseData := serversAsJson

	acceptsGzipEncoding := strings.Contains(request.Header.Get("Accept-Encoding"), "gzip")

	if acceptsGzipEncoding {
		response.Header().Set("Content-Encoding", "gzip")
		responseData = gzipCompress(responseData)
	}

	response.Write(responseData)
}

func getHttpCacheClient() *cache.Client {
	memcached, _ := memory.NewAdapter(
		memory.AdapterWithAlgorithm(memory.LRU),
		memory.AdapterWithCapacity(10000),
	)
	cacheClient, _ := cache.NewClient(
		cache.ClientWithAdapter(memcached),
		cache.ClientWithTTL(5*time.Second),
	)

	return cacheClient
}
