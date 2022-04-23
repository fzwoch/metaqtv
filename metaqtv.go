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
	jsonOut := newMutexStore()

	go func() {
		quakeServers := make(map[SocketAddress]QuakeServer)
		ticker := time.NewTicker(time.Duration(conf.updateInterval) * time.Second)

		for ; true; <-ticker.C {
			var (
				wg    sync.WaitGroup
				mutex sync.Mutex
			)

			quakeServerAddresses := ReadMasterServers(masters, conf.retries, conf.timeout)

			for _, serverAddress := range quakeServerAddresses {
				wg.Add(1)

				go func(serverAddress SocketAddress) {
					defer wg.Done()

					qserver, err := GetServerInfo(serverAddress, conf.retries, conf.timeout, conf.keepalive)

					if err != nil {
						return
					}

					mutex.Lock()
					quakeServers[serverAddress] = qserver
					mutex.Unlock()
				}(serverAddress)
			}

			wg.Wait()

			wg.Add(1)
			go func() {
				defer wg.Done()

				jsonServers := make([]QuakeServer, 0)

				for key, server := range quakeServers {
					if server.keepaliveCount <= 0 {
						delete(quakeServers, key)
						continue
					}

					server.keepaliveCount--

					jsonServers = append(jsonServers, server)

					quakeServers[key] = server
				}

				jsonTmp, err := json.MarshalIndent(jsonServers, "", "\t")
				panicIf(err)

				jsonOut.Write(jsonTmp)
			}()

			wg.Wait()
		}
	}()

	http.HandleFunc("/api/v3/servers", getApiCallback(jsonOut))

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
