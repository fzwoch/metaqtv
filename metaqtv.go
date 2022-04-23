// Copyright (C) 2019-2022 Florian Zwoch <fzwoch@gmail.com>
//
// see https://github.com/eb/metaqtv
// for jogi - get well soon <3

package main

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"encoding/csv"
	"encoding/json"
	"flag"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

func getMasterServers(filePath string) []MasterServer {
	jsonFile, err := os.ReadFile(filePath)
	panicIf(err)

	var masterServers []MasterServer

	err = json.Unmarshal(jsonFile, &masterServers)
	panicIf(err)

	return masterServers
}

func main() {
	var (
		port           int
		updateInterval int
		timeout        int
		retries        int
		configPath     string
		keepalive      int
	)

	flag.IntVar(&port, "port", 3000, "HTTP listen port")
	flag.IntVar(&updateInterval, "interval", 60, "Update interval in seconds")
	flag.IntVar(&timeout, "timeout", 500, "UDP timeout in milliseconds")
	flag.IntVar(&retries, "retry", 5, "UDP retry count")
	flag.StringVar(&configPath, "config", "master_servers.json", "Master servers file")
	flag.IntVar(&keepalive, "keepalive", 3, "Keep server alive for N tries")
	flag.Parse()

	var masterServers = getMasterServers(configPath)

	jsonOut := newMutexStore()

	go func() {

		allServers := make(map[SocketAddress]Server)

		ticker := time.NewTicker(time.Duration(updateInterval) * time.Second)

		for ; true; <-ticker.C {
			var (
				wg    sync.WaitGroup
				mutex sync.Mutex
			)

			servers := make(map[SocketAddress]struct{})

			bufferMaxSize := 8192
			for _, master := range masterServers {
				wg.Add(1)

				go func(master MasterServer) {
					defer wg.Done()

					masterAddress := master.Hostname + ":" + strconv.Itoa(master.Port)
					conn, err := net.Dial("udp4", masterAddress)

					if err != nil {
						log.Println(err)
						return
					}
					defer conn.Close()

					buffer := make([]byte, bufferMaxSize)
					bufferLength := 0

					for i := 0; i < retries; i++ {
						conn.SetDeadline(timeInFuture(timeout))
						mastersServerStatusSequence := []byte{0x63, 0x0a, 0x00}
						_, err = conn.Write(mastersServerStatusSequence)
						if err != nil {
							log.Println(err)
							return
						}

						conn.SetDeadline(timeInFuture(timeout))
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

					validResponseSequence := []byte{0xff, 0xff, 0xff, 0xff, 0x64, 0x0a}
					actualResponseSequence := buffer[:len(validResponseSequence)]
					isValidResponseSequence := bytes.Equal(actualResponseSequence, validResponseSequence)
					if !isValidResponseSequence {
						log.Println(master.Hostname + ":" + strconv.Itoa(master.Port) + ": Response error")
						return
					}

					reader := bytes.NewReader(buffer[6:bufferLength])

					mutex.Lock()

					for {
						var host SocketAddress

						err = binary.Read(reader, binary.BigEndian, &host)
						if err != nil {
							break
						}

						servers[host] = struct{}{}
					}

					mutex.Unlock()
				}(master)
			}

			wg.Wait()

			for server := range servers {
				wg.Add(1)

				go func(server SocketAddress) {
					defer wg.Done()

					ip := net.IPv4(server.Ip[0], server.Ip[1], server.Ip[2], server.Ip[3])

					conn, err := net.Dial("udp4", ip.String()+":"+strconv.Itoa(int(server.Port)))
					if err != nil {
						log.Println(err)
						return
					}
					defer conn.Close()

					buffer := make([]byte, bufferMaxSize)
					bufferLength := 0

					for i := 0; i < retries; i++ {
						conn.SetDeadline(timeInFuture(timeout))
						qtvServerStatusSequence := []byte{0xff, 0xff, 0xff, 0xff, 's', 't', 'a', 't', 'u', 's', ' ', '3', '2', 0x0a}
						_, err = conn.Write(qtvServerStatusSequence)
						if err != nil {
							log.Println(err)
							return
						}

						conn.SetDeadline(timeInFuture(timeout))
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

					fields, err := reader.Read()
					if err != nil {
						log.Println(err)
						return
					}

					if fields[3] == "" {
						// these are the servers that are not configured correctly,
						// that means they are not reporting their qtv ip as they should.
						return
					}

					for i := 0; i < retries; i++ {
						conn.SetDeadline(timeInFuture(timeout))
						serverStatusSequence := []byte{0xff, 0xff, 0xff, 0xff, 's', 't', 'a', 't', 'u', 's', ' ', '2', '3', 0x0a}
						_, err = conn.Write(serverStatusSequence)
						if err != nil {
							log.Println(err)
							return
						}

						conn.SetDeadline(timeInFuture(timeout))
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
						log.Println(ip.String() + ":" + strconv.Itoa(int(server.Port)) + ": Response error")
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

					server_ := Server{
						Title:          "",
						Description:    "",
						Ip:             ip.String(),
						SocketAddress:  ip.String() + ":" + strconv.Itoa(int(server.Port)),
						Port:           server.Port,
						Settings:       map[string]string{},
						Players:        make([]Player, 0),
						Spectators:     make([]Spectator, 0),
						QTV:            make([]QTV, 0),
						keepaliveCount: keepalive,
					}

					server_.QTV = append(server_.QTV, QTV{
						Host:       fields[2],
						Address:    fields[3],
						Spectators: make([]string, 0),
					})

					for i := 0; i < len(settings)-1; i += 2 {
						server_.Settings[settings[i]] = settings[i+1]
					}

					if val, ok := server_.Settings["hostname"]; ok {
						server_.Settings["hostname"] = quakeTextToPlainText(val)
						server_.Title = server_.Settings["hostname"]
					}
					if val, ok := server_.Settings["map"]; ok {
						server_.Map = val
					}
					if val, ok := server_.Settings["maxclients"]; ok {
						value, _ := strconv.Atoi(val)
						server_.MaxPlayers = value
					}
					if val, ok := server_.Settings["maxspectators"]; ok {
						value, _ := strconv.Atoi(val)
						server_.MaxSpectators = value
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
							server_.Spectators = append(server_.Spectators, Spectator{
								Name:    client.Name,
								NameInt: client.NameInt,
								IsBot:   client.IsBot,
							})
						} else {
							server_.Players = append(server_.Players, client.Player)
						}
					}

					mutex.Lock()
					allServers[server] = server_
					mutex.Unlock()
				}(server)
			}

			wg.Wait()

			wg.Add(1)
			go func() {
				defer wg.Done()

				jsonServers := make([]Server, 0)

				for key, server := range allServers {
					if server.keepaliveCount <= 0 {
						delete(allServers, key)
						continue
					}

					server.keepaliveCount--

					jsonServers = append(jsonServers, server)

					allServers[key] = server
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
	err = http.ListenAndServe(":"+strconv.Itoa(port), nil)
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
