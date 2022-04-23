// Copyright (C) 2019-2022 Florian Zwoch <fzwoch@gmail.com>
//
// see https://github.com/eb/metaqtv
// for jogi - get well soon <3

package main

import (
	"bufio"
	"bytes"
	"encoding/csv"
	"encoding/json"
	"log"
	"net"
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

					conn, err := net.Dial("udp4", serverAddress.toString())
					if err != nil {
						log.Println(err)
						return
					}
					defer conn.Close()

					buffer := make([]byte, bufferMaxSize)
					bufferLength := 0

					for i := 0; i < conf.retries; i++ {
						conn.SetDeadline(timeInFuture(conf.timeout))
						qtvServerStatusSequence := []byte{0xff, 0xff, 0xff, 0xff, 's', 't', 'a', 't', 'u', 's', ' ', '3', '2', 0x0a}
						_, err = conn.Write(qtvServerStatusSequence)
						if err != nil {
							log.Println(err)
							return
						}

						conn.SetDeadline(timeInFuture(conf.timeout))
						bufferLength, err = conn.Read(buffer)
						if err != nil {
							continue
						}

						break
					}

					if err != nil {
						// no logging here. it seems that servers may not reply if they do not support
						// this specific "32" status request.
						return
					}

					expectedQtvStatusResponse := []byte{0xff, 0xff, 0xff, 0xff, 'n', 'q', 't', 'v'}
					actualQtvStatusResponse := buffer[:len(expectedQtvStatusResponse)]
					isCorrectQtvResponse := bytes.Equal(actualQtvStatusResponse, expectedQtvStatusResponse)
					if !isCorrectQtvResponse {
						// some servers react to the specific "32" status message but will send the regular
						// status message because they misunderstood our command.
						return
					}

					reader := csv.NewReader(strings.NewReader(string(buffer[5:bufferLength])))
					reader.Comma = ' '

					qtvFields, err := reader.Read()
					if err != nil {
						log.Println(err)
						return
					}

					if qtvFields[3] == "" {
						// these are the servers that are not configured correctly,
						// that means they are not reporting their qtv ip as they should.
						return
					}

					for i := 0; i < conf.retries; i++ {
						conn.SetDeadline(timeInFuture(conf.timeout))
						serverStatusSequence := []byte{0xff, 0xff, 0xff, 0xff, 's', 't', 'a', 't', 'u', 's', ' ', '2', '3', 0x0a}
						_, err = conn.Write(serverStatusSequence)
						if err != nil {
							log.Println(err)
							return
						}

						conn.SetDeadline(timeInFuture(conf.timeout))
						bufferLength, err = conn.Read(buffer)
						if err != nil {
							continue
						}

						break
					}

					if err != nil {
						log.Println(err)
						return
					}

					expectedStatusResponse := []byte{0xff, 0xff, 0xff, 0xff, 'n', '\\'}
					actualStatusResponse := buffer[:len(expectedStatusResponse)]
					isCorrectResponse := bytes.Equal(actualStatusResponse, expectedStatusResponse)
					if !isCorrectResponse {
						log.Println(serverAddress.toString() + ": Response error")
						return
					}

					scanner := bufio.NewScanner(strings.NewReader(string(buffer[6:bufferLength])))
					scanner.Scan()

					settings := strings.FieldsFunc(scanner.Text(), func(r rune) bool {
						if r == '\\' {
							return true
						}
						return false
					})

					qserver := newQuakeServer()
					qserver.Address = serverAddress.toString()
					qserver.keepaliveCount = conf.keepalive

					qserver.QTV = append(qserver.QTV, QtvServer{
						Host:       qtvFields[2],
						Address:    qtvFields[3],
						Spectators: make([]string, 0),
					})

					for i := 0; i < len(settings)-1; i += 2 {
						qserver.Settings[settings[i]] = settings[i+1]
					}

					if val, ok := qserver.Settings["hostname"]; ok {
						qserver.Settings["hostname"] = quakeTextToPlainText(val)
						qserver.Title = qserver.Settings["hostname"]
					}
					if val, ok := qserver.Settings["map"]; ok {
						qserver.Map = val
					}
					if val, ok := qserver.Settings["maxclients"]; ok {
						value, _ := strconv.Atoi(val)
						qserver.MaxPlayers = value
					}
					if val, ok := qserver.Settings["maxspectators"]; ok {
						value, _ := strconv.Atoi(val)
						qserver.MaxSpectators = value
					}

					for scanner.Scan() {
						reader := csv.NewReader(strings.NewReader(scanner.Text()))
						reader.Comma = ' '

						clientRecord, err := reader.Read()
						if err != nil {
							log.Println(err)
							return
						}

						client, err := parseClientRecord(clientRecord)
						if err != nil {
							continue
						}

						if client.IsSpec {
							qserver.Spectators = append(qserver.Spectators, Spectator{
								Name:    client.Name,
								NameInt: client.NameInt,
								IsBot:   client.IsBot,
							})
						} else {
							qserver.Players = append(qserver.Players, client.Player)
						}
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
