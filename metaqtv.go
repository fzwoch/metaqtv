// Copyright (C) 2019 Florian Zwoch <fzwoch@gmail.com>
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
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

var charset = [...]byte{
	' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ',
	'[', ']', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9', ' ', ' ', ' ', ' ',
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

	jsonFile, err := ioutil.ReadFile(config)
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

	jsonOut := struct {
		m sync.RWMutex
		b []byte
	}{
		b: []byte("{}"),
	}

	go func() {
		type host struct {
			IP   [4]byte
			Port uint16
		}

		type serverItem struct {
			Hostname       string
			IPAddress      string `json:"IpAddress"`
			Port           uint16
			Link           string
			keepaliveCount int
			Players        []struct {
				Name string
			}
		}

		allServers := make(map[host]serverItem)

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

					qtv := serverItem{
						Hostname:       ip.String(),
						IPAddress:      ip.String(),
						Port:           server.Port,
						Link:           "http://" + strings.TrimLeft(strings.TrimLeft(fields[3], "1234567890"), "@") + "/watch.qtv?sid=" + strings.Split(fields[3], "@")[0],
						keepaliveCount: keepalive,
						Players: make([]struct {
							Name string
						}, 0),
					}

					scanner := bufio.NewScanner(strings.NewReader(string(data[6:s])))

					scanner.Scan()

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

						player[4] = strings.TrimPrefix(player[4], "\\s\\")

						if player[4] == "[ServeMe]" {
							continue
						}

						name := []byte(player[4])

						for i := range name {
							name[i] &= 0x7f

							if name[i] < byte(len(charset)) {
								name[i] = charset[name[i]]
							}
						}

						qtv.Players = append(qtv.Players, struct {
							Name string
						}{Name: strings.TrimSpace(string(name))})
					}

					m.Lock()
					allServers[server] = qtv
					m.Unlock()
				}(server)
			}

			wg.Wait()

			jsonServers := struct {
				Servers [1]struct {
					GameStates []serverItem
				}
				ServerCount   int
				PlayerCount   int
				ObserverCount int
			}{}

			for key, server := range allServers {
				if server.keepaliveCount <= 0 {
					delete(allServers, key)
					continue
				}

				server.keepaliveCount--

				jsonServers.PlayerCount += len(server.Players)
				jsonServers.Servers[0].GameStates = append(jsonServers.Servers[0].GameStates, server)

				allServers[key] = server
			}

			jsonServers.ServerCount = len(jsonServers.Servers[0].GameStates)

			jsonTmp, err := json.MarshalIndent(jsonServers, "", "\t")
			if err != nil {
				panic(err)
			}

			jsonOut.m.Lock()
			jsonOut.b = jsonTmp
			jsonOut.m.Unlock()
		}
	}()

	http.HandleFunc("/api/v1/servers", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		jsonOut.m.RLock()
		w.Write(jsonOut.b)
		jsonOut.m.RUnlock()
	})

	err = http.ListenAndServe(":"+strconv.Itoa(port), nil)
	if err != nil {
		panic(err)
	}
}
