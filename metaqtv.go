// Copyright (C) 2019-2022 Florian Zwoch <fzwoch@gmail.com>
//
// see https://github.com/eb/metaqtv
// for jogi - get well soon <3

package main

import (
	"encoding/json"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

func getMasterServersFromJsonFile(filePath string) []SocketAddress {
	jsonFile, err := os.ReadFile(filePath)
	panicIf(err)

	var result []SocketAddress
	err = json.Unmarshal(jsonFile, &result)
	panicIf(err)

	return result
}

const bufferMaxSize = 8192

func main() {
	conf := getConfig()
	masters := getMasterServersFromJsonFile(conf.masterServersFile)
	responseJsonData := newMutexStore()

	go func() {
		ticker := time.NewTicker(time.Duration(conf.updateInterval) * time.Second)

		for ; true; <-ticker.C {
			wg := sync.WaitGroup{}
			wg.Add(1)

			go func() {
				defer wg.Done()

				quakeServerAddresses := ReadMasterServers(masters, conf.retries, conf.timeout)
				quakeServers := ReadServers(quakeServerAddresses, conf.retries, conf.timeout)

				serversAsJson, err := json.MarshalIndent(quakeServers, "", "\t")
				panicIf(err)

				responseJsonData.Write(serversAsJson)
			}()

		}
	}()

	http.HandleFunc("/api/v3/servers", getApiCallback(responseJsonData))

	var err error
	err = http.ListenAndServe(":"+strconv.Itoa(conf.httpPort), nil)
	panicIf(err)
}

func getApiCallback(store *MutexStore) func(response http.ResponseWriter, request *http.Request) {
	return func(response http.ResponseWriter, request *http.Request) {
		response.Header().Set("Content-Type", "application/json")
		responseData := store.Read()

		acceptsGzipEncoding := strings.Contains(request.Header.Get("Accept-Encoding"), "gzip")

		if acceptsGzipEncoding {
			response.Header().Set("Content-Encoding", "gzip")
			responseData = gzipCompress(responseData)
		}

		_, err := response.Write(responseData)
		panicIf(err)
	}
}
