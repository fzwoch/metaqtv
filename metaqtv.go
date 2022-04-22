// Copyright (C) 2019-2022 Florian Zwoch <fzwoch@gmail.com>
//
// see https://github.com/eb/metaqtv
// for jogi - get well soon <3

package main

import (
	"bufio"
	"bytes"
	"compress/gzip"
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

type serverItemV1 struct {
	Hostname  string
	IPAddress string `json:"IpAddress"`
	Port      uint16
	Link      string
	Players   []struct {
		Name string
	}

	keepaliveCount int
}

type serverItemV2 struct {
	IPAddress     string `json:"IpAddress"`
	Address       string
	Description   string
	Title         string
	Port          uint16
	Country       string
	IsProxy       bool
	Map           string
	MaxClients    int
	MaxSpectators int
	Settings      map[string]string
	QtvOnly       bool
	Players       []struct {
		Colors  [2]int
		Frags   int
		Ping    int
		Spec    bool
		Name    string
		NameRaw []int
		Team    string
		TeamRaw []int
		Time    int
		IsBot   bool
		Flag    string
	}
	QTV []struct {
		Host     string
		Address  string
		Specs    int
		SpecList []string
	}

	keepaliveCount int
}

type MutexStore struct {
	sync.RWMutex
	data []byte
}

func (store *MutexStore) Write(data []byte) {
	store.Lock()
	store.data = data
	store.Unlock()
}

func (store *MutexStore) Read() []byte {
	store.RLock()
	data := store.data
	store.RUnlock()
	return data
}

func newMutexStore() *MutexStore {
	store := MutexStore{data: make([]byte, 0)}
	return &store
}

func panicIf(err error) {
	if err != nil {
		panic(err)
	}
}

const ColFrags int = 1
const ColTime int = 2
const ColPing int = 3
const ColName int = 4
const ColColorTop int = 6
const ColColorBottom int = 7
const ColTeam int = 8

func main() {
	var (
		port           int
		updateInterval int
		timeout        int
		retries        int
		config         string
		keepalive      int
	)

	flag.IntVar(&port, "port", 3000, "HTTP listen port")
	flag.IntVar(&updateInterval, "interval", 60, "Update interval in seconds")
	flag.IntVar(&timeout, "timeout", 500, "UDP timeout in milliseconds")
	flag.IntVar(&retries, "retry", 5, "UDP retry count")
	flag.StringVar(&config, "config", "metaqtv.json", "Master server config file")
	flag.IntVar(&keepalive, "keepalive", 3, "Keep server alive for N tries")
	flag.Parse()

	jsonFile, err := os.ReadFile(config)
	panicIf(err)

	type masterServer struct {
		Hostname string `json:"hostname"`
		Port     int    `json:"port"`
	}

	var masterServers []masterServer

	err = json.Unmarshal(jsonFile, &masterServers)
	panicIf(err)

	jsonOutV1 := newMutexStore()
	jsonOutV2 := newMutexStore()

	go func() {
		type host struct {
			IP   [4]byte
			Port uint16
		}

		allServersV1 := make(map[host]serverItemV1)
		allServersV2 := make(map[host]serverItemV2)

		ticker := time.NewTicker(time.Duration(updateInterval) * time.Second)

		for ; true; <-ticker.C {
			var (
				wg    sync.WaitGroup
				mutex sync.Mutex
			)

			servers := make(map[host]struct{})

			bufferMaxSize := 8192
			for _, master := range masterServers {
				wg.Add(1)

				go func(master masterServer) {
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
						conn.SetDeadline(time.Now().Add(time.Duration(timeout) * time.Millisecond))
						_, err = conn.Write([]byte{0x63, 0x0a, 0x00})
						if err != nil {
							log.Println(err)
							return
						}

						conn.SetDeadline(time.Now().Add(time.Duration(timeout) * time.Millisecond))
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

					if !bytes.Equal(buffer[:6], []byte{0xff, 0xff, 0xff, 0xff, 0x64, 0x0a}) {
						log.Println(master.Hostname + ":" + strconv.Itoa(master.Port) + ": Response error")
						return
					}

					reader := bytes.NewReader(buffer[6:bufferLength])

					mutex.Lock()

					for {
						var host host

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

				go func(server host) {
					defer wg.Done()

					ip := net.IPv4(server.IP[0], server.IP[1], server.IP[2], server.IP[3])

					conn, err := net.Dial("udp4", ip.String()+":"+strconv.Itoa(int(server.Port)))
					if err != nil {
						log.Println(err)
						return
					}
					defer conn.Close()

					buffer := make([]byte, bufferMaxSize)
					bufferLength := 0

					for i := 0; i < retries; i++ {
						conn.SetDeadline(time.Now().Add(time.Duration(timeout) * time.Millisecond))
						_, err = conn.Write([]byte{0xff, 0xff, 0xff, 0xff, 's', 't', 'a', 't', 'u', 's', ' ', '3', '2', 0x0a})
						if err != nil {
							log.Println(err)
							return
						}

						conn.SetDeadline(time.Now().Add(time.Duration(timeout) * time.Millisecond))
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

					if !bytes.Equal(buffer[:8], []byte{0xff, 0xff, 0xff, 0xff, 'n', 'q', 't', 'v'}) {
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
						conn.SetDeadline(time.Now().Add(time.Duration(timeout) * time.Millisecond))
						_, err = conn.Write([]byte{0xff, 0xff, 0xff, 0xff, 's', 't', 'a', 't', 'u', 's', ' ', '2', '3', 0x0a})
						if err != nil {
							log.Println(err)
							return
						}

						conn.SetDeadline(time.Now().Add(time.Duration(timeout) * time.Millisecond))
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

					if !bytes.Equal(buffer[:6], []byte{0xff, 0xff, 0xff, 0xff, 'n', '\\'}) {
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

					qtv := serverItemV1{
						Hostname:  ip.String(),
						IPAddress: ip.String(),
						Port:      server.Port,
						Link:      "http://" + strings.TrimLeft(strings.TrimLeft(fields[3], "1234567890"), "@") + "/watch.qtv?sid=" + strings.Split(fields[3], "@")[0],
						Players: make([]struct {
							Name string
						}, 0),
						keepaliveCount: keepalive,
					}

					qtvV2 := serverItemV2{
						IPAddress:   ip.String(),
						Address:     ip.String() + ":" + strconv.Itoa(int(server.Port)),
						Description: "",
						Title:       "",
						Port:        server.Port,
						Settings:    map[string]string{},
						Players: make([]struct {
							Colors  [2]int
							Frags   int
							Ping    int
							Spec    bool
							Name    string
							NameRaw []int
							Team    string
							TeamRaw []int
							Time    int
							IsBot   bool
							Flag    string
						}, 0),
						QTV: make([]struct {
							Host     string
							Address  string
							Specs    int
							SpecList []string
						}, 0),
						keepaliveCount: keepalive,
					}

					qtvV2.QTV = append(qtvV2.QTV, struct {
						Host     string
						Address  string
						Specs    int
						SpecList []string
					}{
						Host:     fields[2],
						Address:  fields[3],
						SpecList: make([]string, 0),
					})

					for i := 0; i < len(settings)-1; i += 2 {
						qtvV2.Settings[settings[i]] = settings[i+1]
					}

					if val, ok := qtvV2.Settings["hostname"]; ok {
						qtvV2.Settings["hostname"] = quakeTextToPlainText(val)
						qtvV2.Title = qtvV2.Settings["hostname"]
					}
					if val, ok := qtvV2.Settings["map"]; ok {
						qtvV2.Map = val
					}
					if val, ok := qtvV2.Settings["maxclients"]; ok {
						n, _ := strconv.Atoi(val)
						qtvV2.MaxClients = n
					}
					if val, ok := qtvV2.Settings["maxspectators"]; ok {
						n, _ := strconv.Atoi(val)
						qtvV2.MaxSpectators = n
					}

					for scanner.Scan() {
						reader := csv.NewReader(strings.NewReader(scanner.Text()))
						reader.Comma = ' '

						player, err := reader.Read()
						if err != nil {
							log.Println(err)
							return
						}

						expectedPlayerColumnCount := 9

						if len(player) != expectedPlayerColumnCount {
							continue
						}

						nameRawStr := player[ColName]
						if strings.HasSuffix(nameRawStr, "[ServeMe]") {
							continue
						}

						var isSpec bool
						spectatorPrefix := "\\s\\"
						if strings.HasPrefix(nameRawStr, spectatorPrefix) {
							nameRawStr = strings.TrimPrefix(nameRawStr, spectatorPrefix)
							isSpec = true
						}

						var (
							nameInt []int
							teamRaw []int
						)

						nameStr := quakeTextToPlainText(nameRawStr)
						nameInt = stringToIntArray(nameStr)

						teamStr := quakeTextToPlainText(player[ColTeam])
						teamRaw = stringToIntArray(teamStr)

						qtv.Players = append(qtv.Players, struct {
							Name string
						}{
							Name: nameStr,
						})

						frags, _ := strconv.Atoi(player[ColFrags])
						time_, _ := strconv.Atoi(player[ColTime])
						ping, _ := strconv.Atoi(player[ColPing])
						colorTop, _ := strconv.Atoi(player[ColColorTop])
						colorBottom, _ := strconv.Atoi(player[ColColorBottom])

						if isSpec {
							ping = -ping
						}

						qtvV2.Players = append(qtvV2.Players, struct {
							Colors  [2]int
							Frags   int
							Ping    int
							Spec    bool
							Name    string
							NameRaw []int
							Team    string
							TeamRaw []int
							Time    int
							IsBot   bool
							Flag    string
						}{
							Colors:  [2]int{colorTop, colorBottom},
							Frags:   frags,
							Ping:    ping,
							Spec:    isSpec,
							Name:    nameStr,
							NameRaw: nameInt,
							Team:    teamStr,
							TeamRaw: teamRaw,
							Time:    time_,
						})
					}

					mutex.Lock()
					allServersV1[server] = qtv
					allServersV2[server] = qtvV2
					mutex.Unlock()
				}(server)
			}

			wg.Wait()

			wg.Add(1)
			go func() {
				defer wg.Done()

				jsonServersV2 := make([]serverItemV2, 0)

				for key, server := range allServersV2 {
					if server.keepaliveCount <= 0 {
						delete(allServersV2, key)
						continue
					}

					server.keepaliveCount--

					jsonServersV2 = append(jsonServersV2, server)

					allServersV2[key] = server
				}

				jsonTmp, err := json.MarshalIndent(jsonServersV2, "", "\t")
				panicIf(err)

				jsonOutV2.Write(jsonTmp)
			}()

			jsonServers := struct {
				Servers [1]struct {
					GameStates []serverItemV1
				}
				ServerCount       int
				ActiveServerCount int
				PlayerCount       int
				ObserverCount     int
			}{
				ObserverCount: -1,
			}

			for key, server := range allServersV1 {
				if server.keepaliveCount <= 0 {
					delete(allServersV1, key)
					continue
				}

				server.keepaliveCount--

				jsonServers.PlayerCount += len(server.Players)
				if len(server.Players) > 0 {
					jsonServers.ActiveServerCount++
				}
				jsonServers.Servers[0].GameStates = append(jsonServers.Servers[0].GameStates, server)

				allServersV1[key] = server
			}

			jsonServers.ServerCount = len(jsonServers.Servers[0].GameStates)

			jsonTmp, err := json.MarshalIndent(jsonServers, "", "\t")
			panicIf(err)

			jsonOutV1.Write(jsonTmp)

			wg.Wait()
		}
	}()

	http.HandleFunc("/api/v1/servers", getApiCallback(jsonOutV1))
	http.HandleFunc("/api/v2/servers", getApiCallback(jsonOutV2))

	err = http.ListenAndServe(":"+strconv.Itoa(port), nil)
	panicIf(err)
}

func gzipCompress(data []byte) []byte {
	buffer := bytes.NewBuffer(make([]byte, 0))
	writer := gzip.NewWriter(buffer)
	_, err := writer.Write(data)
	panicIf(err)

	err = writer.Close()
	panicIf(err)

	return buffer.Bytes()
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

func quakeTextToPlainText(value string) string {
	readableTextBytes := []byte(value)

	var charset = [...]byte{
		' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ',
		'[', ']', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9', ' ', ' ', ' ', ' ',
	}

	for i := range value {
		readableTextBytes[i] &= 0x7f

		if value[i] < byte(len(charset)) {
			readableTextBytes[i] = charset[value[i]]
		}
	}

	return strings.TrimSpace(string(readableTextBytes))
}

func stringToIntArray(value string) []int {
	intText := make([]int, len(value))

	for i := range value {
		intText[i] = int(value[i])
	}

	return intText
}
