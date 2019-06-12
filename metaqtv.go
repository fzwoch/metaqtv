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
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

type masterServer struct {
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

	var masterServers []masterServer

	err = json.Unmarshal(jsonFile, &masterServers)
	if err != nil {
		panic(err)
	}

	jsonOut := jsonOut{
		b: []byte("{}"),
	}

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

			servers := make(map[host]struct{})

			for _, master := range masterServers {
				wg.Add(1)

				go func(master masterServer) {
					defer wg.Done()

					c, err := net.Dial("udp", master.Hostname+":"+strconv.Itoa(master.Port))
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

					binary.Read(r, binary.BigEndian, &tmp)
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

						servers[host] = struct{}{}
					}

					m.Unlock()
				}(master)
			}

			wg.Wait()

			qtvServers := make(map[host]struct{})

			for server := range servers {
				wg.Add(1)

				go func(server host) {
					defer wg.Done()

					ip := net.IPv4(server.IP[0], server.IP[1], server.IP[2], server.IP[3])

					c, err := net.Dial("udp", ip.String()+":"+strconv.Itoa(int(server.Port)))
					if err != nil {
						log.Println(err)
						return
					}
					defer c.Close()

					data := make([]byte, 4096)

					var s int
					for i := 0; ; {
						c.SetDeadline(time.Now().Add(time.Second))
						_, err = c.Write([]byte{0xff, 0xff, 0xff, 0xff, 's', 't', 'a', 't', 'u', 's', 0x0a})
						if err != nil {
							log.Println(err)
							return
						}

						c.SetDeadline(time.Now().Add(time.Second))
						s, err = c.Read(data)
						if err != nil {
							i++
							if i > timeout {
								log.Println(err)
								return
							}
							continue
						}
						break
					}

					bla := strings.Split(string(data[:s]), "\\")

					for i := 1; i < len(bla); i += 2 {
						if bla[i] == "*version" {
							if strings.HasPrefix(bla[i+1], "QTV") {
								m.Lock()
								qtvServers[server] = struct{}{}
								m.Unlock()
							} else {
								for i := 0; ; {
									c.SetDeadline(time.Now().Add(time.Second))
									_, err = c.Write([]byte{0xff, 0xff, 0xff, 0xff, 's', 't', 'a', 't', 'u', 's', ' ', '3', '2', 0x0a})
									if err != nil {
										log.Println(err)
										return
									}

									c.SetDeadline(time.Now().Add(time.Second))
									s, err = c.Read(data)
									if err != nil {
										i++
										if i > timeout {
											log.Println(err)
											return
										}
										continue
									}

									r := regexp.MustCompile("\".*?\"|\\S+")
									bla := r.FindAllString(string(data[5:s]), -1)

									if len(bla) > 3 && bla[0] == "qtv" && bla[3] != "\"\"" {
										x := strings.Trim(bla[3], "\"")
										x = strings.TrimLeft(x, "1234567890@")
										y := strings.Split(x, ":")

										ip, err := net.LookupIP(y[0])
										if err != nil {
											panic(err)
										}

										tmp, err := strconv.Atoi(y[1])

										var h host

										h.IP[0] = ip[0][0]
										h.IP[1] = ip[0][1]
										h.IP[2] = ip[0][2]
										h.IP[3] = ip[0][3]

										h.Port = uint16(tmp)

										//	log.Println(y)

										m.Lock()
										qtvServers[h] = struct{}{}
										m.Unlock()
									}
									break
								}
							}
							break
						}
					}
				}(server)
			}

			wg.Wait()

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

			for server := range qtvServers {
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
