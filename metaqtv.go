// Copyright (C) 2019 Florian Zwoch <fzwoch@gmail.com>
//
// see https://github.com/eb/metaqtv
// for jogi - get well soon <3

package main

import (
	"encoding/json"
	"encoding/xml"
	"flag"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
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
	)

	flag.IntVar(&port, "port", 3000, "HTTP listen port")
	flag.IntVar(&updateInterval, "interval", 60, "Update interval in seconds")
	flag.IntVar(&timeout, "timeout", 5, "RSS request timeout in seconds")
	flag.Parse()

	f, err := os.Open("metaqtv.json")
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

	ticker := time.NewTicker(time.Duration(updateInterval) * time.Second)

	go func() {
		for ; true; <-ticker.C {
			var (
				wg sync.WaitGroup
				m  sync.Mutex
			)

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

			for _, server := range servers {
				wg.Add(1)

				go func(server qtvServer) {
					defer wg.Done()

					c := http.Client{
						Timeout: time.Duration(timeout) * time.Second,
					}

					resp, err := c.Get("http://" + server.Hostname + ":" + strconv.Itoa(server.Port) + "/rss")
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
