package main

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/victorspringer/http-cache"
	"github.com/victorspringer/http-cache/adapter/memory"
)

func serversResponse(servers []QuakeServer, response http.ResponseWriter, request *http.Request) {
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
