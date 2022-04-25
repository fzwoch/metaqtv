package main

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/victorspringer/http-cache"
	"github.com/victorspringer/http-cache/adapter/memory"
)

func httpJsonResponse(data any, response http.ResponseWriter, request *http.Request) {
	response.Header().Set("Content-Type", "application/json")

	responseBody, _ := json.MarshalIndent(data, "", "\t")

	acceptsGzipEncoding := strings.Contains(request.Header.Get("Accept-Encoding"), "gzip")

	if acceptsGzipEncoding {
		response.Header().Set("Content-Encoding", "gzip")
		responseBody = gzipCompress(responseBody)
	}

	response.Write(responseBody)
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

func gzipCompress(data []byte) []byte {
	buffer := bytes.NewBuffer(make([]byte, 0))
	writer := gzip.NewWriter(buffer)
	writer.Write(data)
	writer.Close()

	return buffer.Bytes()
}
