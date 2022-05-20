package mhttp

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/rs/cors"
	"github.com/victorspringer/http-cache"
	"github.com/victorspringer/http-cache/adapter/memory"
)

type HttpServer struct {
	Endpoints Endpoints
}

type Endpoints map[string]http.HandlerFunc

func NewServer() HttpServer {
	return HttpServer{}
}

func (server HttpServer) Serve(port int) {
	isDevelopment := false

	var handler http.Handler
	cacheClient := getCacheClient()

	if isDevelopment {
		mux := http.NewServeMux() // CORS
		for url, handler := range server.Endpoints {
			mux.Handle(url, cacheClient.Middleware(handler))
		}
		handler = cors.Default().Handler(mux) // CORS
	} else {
		for url, handler := range server.Endpoints {
			http.Handle(url, cacheClient.Middleware(handler))
		}
		handler = nil
	}

	// serve
	serverAddress := fmt.Sprintf(":%d", port)
	err := http.ListenAndServe(serverAddress, handler)

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func getCacheClient() *cache.Client {
	memcached, _ := memory.NewAdapter(
		memory.AdapterWithAlgorithm(memory.LRU),
		memory.AdapterWithCapacity(10000),
	)
	cacheClient, _ := cache.NewClient(
		cache.ClientWithAdapter(memcached),
		cache.ClientWithTTL(1*time.Second),
	)

	return cacheClient
}

func HandlerBySource(source func() any) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		responseBody, _ := jsonMarshalNoEscapeHtml(source())
		JsonResponse(responseBody, w, r)
	}
}

func JsonResponse(responseBody []byte, response http.ResponseWriter, request *http.Request) {
	response.Header().Set("Content-Type", "application/json")
	acceptsGzipEncoding := strings.Contains(request.Header.Get("Accept-Encoding"), "gzip")

	if acceptsGzipEncoding {
		response.Header().Set("Content-Encoding", "gzip")
		responseBody = gzipCompress(responseBody)
	}

	response.Write(responseBody)
}

func jsonMarshalNoEscapeHtml(value any) ([]byte, error) {
	var dst bytes.Buffer
	enc := json.NewEncoder(&dst)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "\t")
	err := enc.Encode(value)
	if err != nil {
		return nil, err
	}
	return dst.Bytes(), nil
}

func gzipCompress(data []byte) []byte {
	buffer := bytes.NewBuffer(make([]byte, 0))
	writer := gzip.NewWriter(buffer)
	writer.Write(data)
	writer.Close()

	return buffer.Bytes()
}
