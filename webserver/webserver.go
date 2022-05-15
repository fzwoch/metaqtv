package webserver

import (
	"net/http"
	"time"

	"github.com/victorspringer/http-cache"
	"github.com/victorspringer/http-cache/adapter/memory"
	"metaqtv/geo"
	apiHandler "metaqtv/webserver/handler"
)

func Serve(addr string, serversWithGeo *[]geo.ServerWithGeo) {
	// endpoints
	api := make(map[string]http.HandlerFunc, 0)
	api["/servers"] = apiHandler.Mvdsv(serversWithGeo)
	api["/proxies"] = apiHandler.Qwforwards(serversWithGeo)
	api["/qtv"] = apiHandler.Qtv(serversWithGeo)
	api["/server_to_qtv"] = apiHandler.ServerToQtv(serversWithGeo)
	api["/qtv_to_server"] = apiHandler.QtvToServer(serversWithGeo)

	// middleware
	cacheClient := getCacheClient()
	for url, handler := range api {
		http.Handle(url, cacheClient.Middleware(handler))
	}

	// serve
	http.ListenAndServe(addr, nil)
}

func getCacheClient() *cache.Client {
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
