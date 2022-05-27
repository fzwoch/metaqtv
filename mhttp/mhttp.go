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

	"github.com/victorspringer/http-cache"
	"github.com/victorspringer/http-cache/adapter/memory"
	"metaqtv/dataprovider"
)

type Api struct {
	Provider  *dataprovider.DataProvider
	BaseUrl   string
	Endpoints Endpoints
}

type Endpoints map[string]http.HandlerFunc

func Serve(port int, endpoints Endpoints) {
	// middleware
	cacheClient := getCacheClient()
	for url, handler := range endpoints {
		http.Handle(url, cacheClient.Middleware(handler))
	}

	// serve
	serverAddress := fmt.Sprintf(":%d", port)
	err := http.ListenAndServe(serverAddress, nil)

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
		cache.ClientWithTTL(5*time.Second),
	)

	return cacheClient
}

func CreateHandler(dataSource func() any) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		responseBody, _ := JsonMarshalNoEscapeHtml(dataSource())
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

func JsonMarshalNoEscapeHtml(value any) ([]byte, error) {
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
