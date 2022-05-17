package webserver

import (
	"fmt"
	"net/http"
	"time"

	"github.com/rs/cors"
	"github.com/victorspringer/http-cache"
	"github.com/victorspringer/http-cache/adapter/memory"
	"metaqtv/provider"
	apiHandler "metaqtv/webserver/handler"
)

func Serve(httpPort int, dataSource func() []provider.ServerWithGeo) {
	// endpoints
	api := make(map[string]http.HandlerFunc, 0)
	api["/servers"] = apiHandler.Mvdsv(dataSource)
	api["/proxies"] = apiHandler.Qwforwards(dataSource)
	api["/qtv"] = apiHandler.Qtv(dataSource)
	api["/fortress"] = apiHandler.Fortress(dataSource)
	api["/server_to_qtv"] = apiHandler.ServerToQtv(dataSource)
	api["/qtv_to_server"] = apiHandler.QtvToServer(dataSource)

	// middleware
	mux := http.NewServeMux() // CORS
	cacheClient := getCacheClient()
	for url, handler := range api {
		// http.Handle(url, cacheClient.Middleware(handler))
		mux.Handle(url, cacheClient.Middleware(handler))
	}

	// serve
	handler := cors.Default().Handler(mux) // CORS
	serverAddr := fmt.Sprintf(":%d", httpPort)
	http.ListenAndServe(serverAddr, handler)
}

func getCacheClient() *cache.Client {
	memcached, _ := memory.NewAdapter(
		memory.AdapterWithAlgorithm(memory.LRU),
		memory.AdapterWithCapacity(10000),
	)
	cacheClient, _ := cache.NewClient(
		cache.ClientWithAdapter(memcached),
		cache.ClientWithTTL(3*time.Second),
	)

	return cacheClient
}
