package handler

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"net/http"
	"strings"

	"metaqtv/scrape"
	"metaqtv/transform"
)

func Mvdsv(dataSource func() []transform.MvdsvWithGeo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		JsonResponse(dataSource(), w, r)
	}
}

func Qwforwards(dataSource func() []transform.QwfwdWithGeo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		JsonResponse(dataSource(), w, r)
	}
}

func Qtv(dataSource func() []transform.QtvWithGeo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		qtvServersWithGeo := transform.ToQtvServers(filter.QtvServers(dataSource()))
		JsonResponse(qtvServersWithGeo, w, r)
	}
}

func ServerToQtv(dataSource func() []transform.QtvWithGeo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		JsonResponse(transform.ServerAddressToQtvStreamUrlMap(dataSource()), w, r)
	}
}

func QtvToServer(dataSource func() []scrape.ServerWithGeo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		JsonResponse(transform.QtvStreamUrlToServerAddressMap(dataSource()), w, r)
	}
}

func JsonResponse(data any, response http.ResponseWriter, request *http.Request) {
	response.Header().Set("Content-Type", "application/json")
	responseBody, _ := json.MarshalIndent(data, "", "\t")
	acceptsGzipEncoding := strings.Contains(request.Header.Get("Accept-Encoding"), "gzip")

	if acceptsGzipEncoding {
		response.Header().Set("Content-Encoding", "gzip")
		responseBody = gzipCompress(responseBody)
	}

	response.Write(responseBody)
}

func gzipCompress(data []byte) []byte {
	buffer := bytes.NewBuffer(make([]byte, 0))
	writer := gzip.NewWriter(buffer)
	writer.Write(data)
	writer.Close()

	return buffer.Bytes()
}
