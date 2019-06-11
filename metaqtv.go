// Copyright (C) 2019 Florian Zwoch <fzwoch@gmail.com>
//
// see https://github.com/eb/metaqtv
// for jogi - get well soon <3

package main

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"encoding/xml"
	"flag"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

type qtvServer struct {
	Hostname string `json:"hostname"`
	Port     int    `json:"port"`
}

type player struct {
	Name        string `xml:"name"`
	Team        string `xml:"team" json:"-"`
	Frags       int    `xml:"frags" json:"-"`
	Ping        int    `xml:"ping" json:"-"`
	PL          int    `xml:"pl" json:"-"`
	TopColor    int    `xml:"topcolor" json:"-"`
	BottomColor int    `xml:"bottomcolor" json:"-"`
}

type spectator struct {
	Name string `xml:"name"`
	Ping int    `xml:"ping"`
	PL   int    `xml:"pl"`
}

type xmlItem struct {
	Title         string      `xml:"title" json:"-"`
	Hostname      string      `xml:"hostname" json:"Hostname"`
	IPAddress     string      `json:"IpAddress"`
	Port          int         `xml:"port"`
	Link          string      `xml:"link"`
	Status        string      `xml:"status" json:"-"`
	Map           string      `xml:"map" json:"-"`
	ObserverCount int         `xml:"observercount" json:"-"`
	Players       []player    `xml:"player"`
	Spectators    []spectator `xml:"spectator" json:"-"`
}

type xmlServer struct {
	XMLName xml.Name  `xml:"rss"`
	Items   []xmlItem `xml:"channel>item"`
}

type jsonOut struct {
	m sync.RWMutex
	b []byte
}

func main() {
	var (
		port           int
		updateInterval int
		timeout        int
		config         string
	)

	flag.IntVar(&port, "port", 3000, "HTTP listen port")
	flag.IntVar(&updateInterval, "interval", 60, "Update interval in seconds")
	flag.IntVar(&timeout, "timeout", 5, "RSS request timeout in seconds")
	flag.StringVar(&config, "config", "metaqtv.json", "QTV server config file")
	flag.Parse()

	f, err := os.Open(config)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	jsonFile, err := ioutil.ReadAll(f)
	if err != nil {
		panic(err)
	}

	var servers []qtvServer

	err = json.Unmarshal(jsonFile, &servers)
	if err != nil {
		panic(err)
	}

	var jsonOut jsonOut

	go func() {
		ticker := time.NewTicker(time.Duration(updateInterval) * time.Second)

		for ; true; <-ticker.C {
			var (
				wg sync.WaitGroup
				m  sync.Mutex
			)

			type host struct {
				IP   [4]byte
				Port uint16
			}

			qtv := make(map[host]struct{})

			for _, server := range servers {
				wg.Add(1)

				go func(server qtvServer) {
					defer wg.Done()

					addr, err := net.ResolveUDPAddr("udp", server.Hostname+":"+strconv.Itoa(server.Port))
					if err != nil {
						log.Println(err)
						return
					}

					c, err := net.DialUDP("udp", nil, addr)
					if err != nil {
						log.Println(err)
						return
					}
					defer c.Close()

					_, err = c.Write([]byte{0x63, 0x0a, 0x00})
					if err != nil {
						log.Println(err)
						return
					}

					data := make([]byte, 4096)

					_, err = c.Read(data)
					if err != nil {
						log.Println(err)
						return
					}

					r := bytes.NewReader(data)

					var tmp [6]byte

					binary.Read(r, binary.LittleEndian, &tmp)
					if tmp != [6]byte{0xff, 0xff, 0xff, 0xff, 0x64, 0x0a} {
						log.Println("Response error")
						return
					}

					m.Lock()

					for {
						var host host

						err = binary.Read(r, binary.BigEndian, &host)
						if err != nil {
							break
						}

						// what?
						if host.Port == 0 {
							continue
						}

						qtv[host] = struct{}{}
					}

					m.Unlock()
				}(server)
			}

			wg.Wait()

			qtvs := make([]host, 0)

			for h := range qtv {
				wg.Add(1)

				go func(h host) {
					defer wg.Done()

					ip := net.IPv4(h.IP[0], h.IP[1], h.IP[2], h.IP[3])

					addr, err := net.ResolveUDPAddr("udp", ip.String()+":"+strconv.Itoa(int(h.Port)))
					if err != nil {
						log.Println(err)
						return
					}

					c, err := net.DialUDP("udp", nil, addr)
					if err != nil {
						log.Println(err)
						return
					}
					defer c.Close()

					c.SetDeadline(time.Now().Add(time.Duration(timeout) * time.Second))

					_, err = c.Write([]byte{0xff, 0xff, 0xff, 0xff, 's', 't', 'a', 't', 'u', 's', 0x0a})
					if err != nil {
						log.Println(err)
						return
					}

					data := make([]byte, 4096)

					_, err = c.Read(data)
					if err != nil {
						log.Println(err)
						return
					}

					bla := strings.Split(string(data), "\\")

					for i := 1; i < len(bla); i += 2 {
						if bla[i] == "*version" {
							if strings.HasPrefix(bla[i+1], "QTV") {
								m.Lock()
								qtvs = append(qtvs, h)
								m.Unlock()
							}
							break
						}
					}
				}(h)
			}

			wg.Wait()

			log.Println(len(qtvs))

			allServers := struct {
				Servers []struct {
					GameStates []xmlItem
				}
				ServerCount   int
				PlayerCount   int
				ObserverCount int
			}{
				Servers: make([]struct {
					GameStates []xmlItem
				}, 1),
			}

			for _, server := range qtvs {
				wg.Add(1)

				go func(server host) {
					defer wg.Done()

					ip := net.IPv4(server.IP[0], server.IP[1], server.IP[2], server.IP[3])

					c := http.Client{
						Timeout: time.Duration(timeout) * time.Second,
					}

					resp, err := c.Get("http://" + ip.String() + ":" + strconv.Itoa(int(server.Port)) + "/rss")
					if err != nil {
						log.Println(err)
						return
					}
					defer resp.Body.Close()

					body, err := ioutil.ReadAll(resp.Body)
					if err != nil {
						panic(err)
					}

					var xmlData xmlServer

					err = xml.Unmarshal(body, &xmlData)
					if err != nil {
						log.Println(err)
						return
					}

					for i, s := range xmlData.Items {
						if xmlData.Items[i].Players == nil {
							xmlData.Items[i].Players = make([]player, 0)
						}

						addr, err := net.LookupIP(s.Hostname)
						if err != nil {
							log.Println(err)
							continue
						}
						xmlData.Items[i].IPAddress = addr[0].String()
					}

					m.Lock()
					allServers.Servers[0].GameStates = append(allServers.Servers[0].GameStates, xmlData.Items...)
					m.Unlock()
				}(server)
			}

			wg.Wait()

			allServers.ServerCount += len(allServers.Servers[0].GameStates)

			for _, s := range allServers.Servers[0].GameStates {
				allServers.PlayerCount += len(s.Players)
				allServers.ObserverCount += s.ObserverCount
			}

			jsonTmp, err := json.MarshalIndent(allServers, "", "\t")
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
