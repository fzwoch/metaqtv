package webserver

import (
	"net/http"
	"time"

	"github.com/rs/cors"
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
	api["/fortress"] = apiHandler.Fortress(serversWithGeo)
	api["/server_to_qtv"] = apiHandler.ServerToQtv(serversWithGeo)
	api["/qtv_to_server"] = apiHandler.QtvToServer(serversWithGeo)

	// middleware
	mux := http.NewServeMux() // CORS
	cacheClient := getCacheClient()
	for url, handler := range api {
		// http.Handle(url, cacheClient.Middleware(handler))
		mux.Handle(url, cacheClient.Middleware(handler))
	}

	// serve
	handler := cors.Default().Handler(mux) // CORS
	http.ListenAndServe(addr, handler)
}

func getCacheClient() *cache.Client {
	memcached, _ := memory.NewAdapter(
		memory.AdapterWithAlgorithm(memory.LRU),
		memory.AdapterWithCapacity(10000),
	)
	cacheClient, _ := cache.NewClient(
		cache.ClientWithAdapter(memcached),
		cache.ClientWithTTL(9*time.Second),
	)

	return cacheClient
}
