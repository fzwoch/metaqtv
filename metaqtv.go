// see https://github.com/eb/metaqtv
// for jogi - get well soon <3

package main

import (
	"encoding/json"
	"encoding/xml"
	"flag"
	"io/ioutil"
	"log"
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

type xmlItem struct {
	Title         string `xml:"title" json:"title"`
	Hostname      string `xml:"hostname" json:"hostname"`
	Port          int    `xml:"port" json:"port"`
	Status        string `xml:"status" json:"status"`
	Map           string `xml:"map" json:"map"`
	ObserverCount int    `xml:"observercount" json:"observercount"`
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
	var port int
	flag.IntVar(&port, "port", 3000, "HTTP listen port")
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

	ticker := time.NewTicker(time.Minute)

	go func() {
		for ; true; <-ticker.C {
			var (
				wg         sync.WaitGroup
				m          sync.Mutex
				allServers []xmlItem
			)

			for _, server := range servers {
				wg.Add(1)

				go func(server qtvServer) {
					defer wg.Done()

					c := http.Client{
						Timeout: 5 * time.Second,
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

					m.Lock()
					allServers = append(allServers, xmlData.Items...)
					m.Unlock()
				}(server)
			}

			wg.Wait()

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

	http.HandleFunc("/api/v2/servers", func(w http.ResponseWriter, r *http.Request) {
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
