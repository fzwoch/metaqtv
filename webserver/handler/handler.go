package handler

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"net/http"
	"strings"

	"metaqtv/provider"
	"metaqtv/webserver/handler/filter"
	"metaqtv/webserver/handler/transform"
)

func Mvdsv(dataSource func() []provider.ServerWithGeo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		mvdsvServersWithGeo := transform.ToMvdsvServers(filter.MvdsvServers(dataSource()))
		jsonResponse(mvdsvServersWithGeo, w, r)
	}
}

func Qwforwards(dataSource func() []provider.ServerWithGeo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		qwforwardsWithGeo := transform.ToQwfwds(filter.Qwforwards(dataSource()))
		jsonResponse(qwforwardsWithGeo, w, r)
	}
}

func Fortress(dataSource func() []provider.ServerWithGeo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		qwforwardsWithGeo := filter.ForstressOneServers(dataSource())
		jsonResponse(qwforwardsWithGeo, w, r)
	}
}

func Qtv(dataSource func() []provider.ServerWithGeo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		qtvServersWithGeo := transform.ToQtvServers(filter.QtvServers(dataSource()))
		jsonResponse(qtvServersWithGeo, w, r)
	}
}

func ServerToQtv(dataSource func() []provider.ServerWithGeo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		jsonResponse(transform.ServerAddressToQtvStreamUrlMap(dataSource()), w, r)
	}
}

func QtvToServer(dataSource func() []provider.ServerWithGeo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		jsonResponse(transform.QtvStreamUrlToServerAddressMap(dataSource()), w, r)
	}
}

func jsonResponse(data any, response http.ResponseWriter, request *http.Request) {
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
