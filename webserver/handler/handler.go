package handler

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"net/http"
	"strings"

	"metaqtv/geo"
	"metaqtv/webserver/handler/filter"
	"metaqtv/webserver/handler/transform"
)

func Mvdsv(servers *[]geo.ServerWithGeo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		mvdsvServersWithGeo := transform.ToMvdsvServers(filter.MvdsvServers(*servers))
		jsonResponse(mvdsvServersWithGeo, w, r)
	}
}

func Qwforwards(servers *[]geo.ServerWithGeo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		qwforwardsWithGeo := transform.ToQwfwds(filter.Qwforwards(*servers))
		jsonResponse(qwforwardsWithGeo, w, r)
	}
}

func Fortress(servers *[]geo.ServerWithGeo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		qwforwardsWithGeo := filter.ForstressOneServers(*servers)
		jsonResponse(qwforwardsWithGeo, w, r)
	}
}

func Qtv(servers *[]geo.ServerWithGeo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		qtvServersWithGeo := transform.ToQtvServers(filter.QtvServers(*servers))
		jsonResponse(qtvServersWithGeo, w, r)
	}
}

func ServerToQtv(servers *[]geo.ServerWithGeo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		jsonResponse(transform.ServerAddressToQtvStreamUrlMap(*servers), w, r)
	}
}

func QtvToServer(servers *[]geo.ServerWithGeo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		jsonResponse(transform.QtvStreamUrlToServerAddressMap(*servers), w, r)
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
