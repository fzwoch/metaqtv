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

var charset = [...]byte{
	' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ',
	'[', ']', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9', ' ', ' ', ' ', ' ',
}

type jsonStore struct {
	sync.RWMutex
	b []byte
	z []byte
}

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
	if err != nil {
		panic(err)
	}

	type masterServer struct {
		Hostname string `json:"hostname"`
		Port     int    `json:"port"`
	}

	var masterServers []masterServer

	err = json.Unmarshal(jsonFile, &masterServers)
	if err != nil {
		panic(err)
	}

	b := bytes.NewBuffer(make([]byte, 0))
	w := gzip.NewWriter(b)
	w.Write([]byte("{}"))
	w.Close()

	jsonOutV1 := jsonStore{
		b: []byte("{}"),
		z: b.Bytes(),
	}

	jsonOutV2 := jsonStore{
		b: []byte("{}"),
		z: b.Bytes(),
	}

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
				wg sync.WaitGroup
				m  sync.Mutex
			)

			servers := make(map[host]struct{})

			for _, master := range masterServers {
				wg.Add(1)

				go func(master masterServer) {
					defer wg.Done()

					c, err := net.Dial("udp4", master.Hostname+":"+strconv.Itoa(master.Port))
					if err != nil {
						log.Println(err)
						return
					}
					defer c.Close()

					s := 0
					data := make([]byte, 8192)

					for i := 0; i < retries; i++ {
						c.SetDeadline(time.Now().Add(time.Duration(timeout) * time.Millisecond))
						_, err = c.Write([]byte{0x63, 0x0a, 0x00})
						if err != nil {
							log.Println(err)
							return
						}

						c.SetDeadline(time.Now().Add(time.Duration(timeout) * time.Millisecond))
						s, err = c.Read(data)
						if err != nil {
							continue
						}

						break
					}

					if err != nil {
						log.Println(err)
						return
					}

					if !bytes.Equal(data[:6], []byte{0xff, 0xff, 0xff, 0xff, 0x64, 0x0a}) {
						log.Println(master.Hostname + ":" + strconv.Itoa(master.Port) + ": Response error")
						return
					}

					r := bytes.NewReader(data[6:s])

					m.Lock()

					for {
						var host host

						err = binary.Read(r, binary.BigEndian, &host)
						if err != nil {
							break
						}

						servers[host] = struct{}{}
					}

					m.Unlock()
				}(master)
			}

			wg.Wait()

			for server := range servers {
				wg.Add(1)

				go func(server host) {
					defer wg.Done()

					ip := net.IPv4(server.IP[0], server.IP[1], server.IP[2], server.IP[3])

					c, err := net.Dial("udp4", ip.String()+":"+strconv.Itoa(int(server.Port)))
					if err != nil {
						log.Println(err)
						return
					}
					defer c.Close()

					s := 0
					data := make([]byte, 8192)

					for i := 0; i < retries; i++ {
						c.SetDeadline(time.Now().Add(time.Duration(timeout) * time.Millisecond))
						_, err = c.Write([]byte{0xff, 0xff, 0xff, 0xff, 's', 't', 'a', 't', 'u', 's', ' ', '3', '2', 0x0a})
						if err != nil {
							log.Println(err)
							return
						}

						c.SetDeadline(time.Now().Add(time.Duration(timeout) * time.Millisecond))
						s, err = c.Read(data)
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

					if !bytes.Equal(data[:8], []byte{0xff, 0xff, 0xff, 0xff, 'n', 'q', 't', 'v'}) {
						// some servers react to the specific "32" status message but will send the regular
						// status message because they misunderstood our command.
						return
					}

					r := csv.NewReader(strings.NewReader(string(data[5:s])))
					r.Comma = ' '

					fields, err := r.Read()
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
						c.SetDeadline(time.Now().Add(time.Duration(timeout) * time.Millisecond))
						_, err = c.Write([]byte{0xff, 0xff, 0xff, 0xff, 's', 't', 'a', 't', 'u', 's', ' ', '2', '3', 0x0a})
						if err != nil {
							log.Println(err)
							return
						}

						c.SetDeadline(time.Now().Add(time.Duration(timeout) * time.Millisecond))
						s, err = c.Read(data)
						if err != nil {
							continue
						}

						break
					}

					if err != nil {
						log.Println(err)
						return
					}

					if !bytes.Equal(data[:6], []byte{0xff, 0xff, 0xff, 0xff, 'n', '\\'}) {
						log.Println(ip.String() + ":" + strconv.Itoa(int(server.Port)) + ": Response error")
						return
					}

					scanner := bufio.NewScanner(strings.NewReader(string(data[6:s])))
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
						hostname := []byte(val)

						for i := range hostname {
							hostname[i] &= 0x7f

							if hostname[i] < byte(len(charset)) {
								hostname[i] = charset[hostname[i]]
							}
						}

						qtvV2.Settings["hostname"] = strings.TrimSpace(string(hostname))
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
						r := csv.NewReader(strings.NewReader(scanner.Text()))
						r.Comma = ' '

						player, err := r.Read()
						if err != nil {
							log.Println(err)
							return
						}

						if len(player) != 9 {
							continue
						}

						if strings.HasSuffix(player[4], "[ServeMe]") {
							continue
						}

						var spec bool
						if strings.HasPrefix(player[4], "\\s\\") {
							player[4] = strings.TrimPrefix(player[4], "\\s\\")
							spec = true
						}

						var (
							nameRaw []int
							teamRaw []int
						)

						name := []byte(player[4])

						for i := range name {
							nameRaw = append(nameRaw, int(name[i]))

							name[i] &= 0x7f

							if name[i] < byte(len(charset)) {
								name[i] = charset[name[i]]
							}
						}

						team := []byte(player[8])

						for i := range team {
							teamRaw = append(teamRaw, int(team[i]))

							team[i] &= 0x7f

							if team[i] < byte(len(charset)) {
								team[i] = charset[team[i]]
							}
						}

						qtv.Players = append(qtv.Players, struct {
							Name string
						}{
							Name: strings.TrimSpace(string(name)),
						})

						frags, _ := strconv.Atoi(player[1])
						time_, _ := strconv.Atoi(player[2])
						ping, _ := strconv.Atoi(player[3])
						colorTop, _ := strconv.Atoi(player[6])
						colorBottom, _ := strconv.Atoi(player[7])

						if spec {
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
							Spec:    spec,
							Name:    strings.TrimSpace(string(name)),
							NameRaw: nameRaw,
							Team:    strings.TrimSpace(string(team)),
							TeamRaw: teamRaw,
							Time:    time_,
						})
					}

					m.Lock()
					allServersV1[server] = qtv
					allServersV2[server] = qtvV2
					m.Unlock()
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
				if err != nil {
					panic(err)
				}

				b := bytes.NewBuffer(make([]byte, 0))
				w := gzip.NewWriter(b)
				w.Write(jsonTmp)
				w.Close()

				jsonOutV2.Lock()
				jsonOutV2.b = jsonTmp
				jsonOutV2.z = b.Bytes()
				jsonOutV2.Unlock()
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
			if err != nil {
				panic(err)
			}

			b := bytes.NewBuffer(make([]byte, 0))
			w := gzip.NewWriter(b)
			w.Write(jsonTmp)
			w.Close()

			jsonOutV1.Lock()
			jsonOutV1.b = jsonTmp
			jsonOutV1.z = b.Bytes()
			jsonOutV1.Unlock()

			wg.Wait()
		}
	}()

	http.HandleFunc("/api/v1/servers", getServerApiRequestCallback(&jsonOutV1))
	http.HandleFunc("/api/v2/servers", getServerApiRequestCallback(&jsonOutV2))

	err = http.ListenAndServe(":"+strconv.Itoa(port), nil)
	if err != nil {
		panic(err)
	}
}

func gzipCompress(data []byte) []byte {
	buffer := bytes.NewBuffer(make([]byte, 0))
	writer := gzip.NewWriter(buffer)
	writer.Write(data)
	writer.Close()
	return buffer.Bytes()
}

func getApiCallback(store *jsonStore) func(w http.ResponseWriter, r *http.Request) {
	return func(writer http.ResponseWriter, req *http.Request) {
		writer.Header().Set("Content-Type", "application/json")

		acceptsGzipEncoding := strings.Contains(req.Header.Get("Accept-Encoding"), "gzip")

		store.RLock()

		if acceptsGzipEncoding {
			writer.Header().Set("Content-Encoding", "gzip")
			writer.Write(store.z)
		} else {
			writer.Write(store.b)
		}

		store.RUnlock()
	}
}
