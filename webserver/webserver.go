package webserver

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/victorspringer/http-cache"
	"github.com/victorspringer/http-cache/adapter/memory"
	"github.com/vikpe/serverstat/qserver"
	"metaqtv/filter"
	"metaqtv/geo"
	"metaqtv/transform"
)

func Serve(addr string, servers *[]qserver.GenericServer) {
	// get geo data
	geoDatabase, err := geo.New()

	if err != nil {
		log.Println("Unable to download geo data.json")
		os.Exit(1)
	}

	handlerByFilter := func(filterFunc func([]qserver.GenericServer) []qserver.GenericServer) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			httpJsonResponse(appendGeo(geoDatabase, filterFunc(*servers)), w, r)
		}
	}

	handlerByMapping := func(mapFunc func([]qserver.GenericServer) map[string]string) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) { httpJsonResponse(mapFunc(*servers), w, r) }
	}

	api := make(map[string]http.HandlerFunc, 0)
	api["/servers"] = handlerByFilter(filter.GameServers)
	api["/proxies"] = handlerByFilter(filter.ProxyServers)
	api["/qtv"] = handlerByFilter(filter.QtvServers)
	api["/server_to_qtv"] = handlerByMapping(transform.ServerAddressToQtvMap)
	api["/qtv_to_server"] = handlerByMapping(transform.QtvToServerAddressMap)

	cacheClient := getHttpCacheClient()
	for url, handler := range api {
		http.Handle(url, cacheClient.Middleware(handler))
	}

	http.ListenAndServe(addr, nil)
}

type ServerWithGeo struct {
	qserver.GenericServer
	Geo geo.Info
}

func appendGeo(geoDb geo.Database, servers []qserver.GenericServer) []ServerWithGeo {
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
