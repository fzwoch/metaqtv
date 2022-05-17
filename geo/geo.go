package geo

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
)

type Info struct {
	CC      string
	Country string
	Region  string
}

type Database map[string]Info

func (db Database) Get(ip string) Info {
	if _, ok := db[ip]; ok {
		return db[ip]
	} else {
		return Info{
			CC:      "",
			Country: "",
			Region:  "",
		}
	}
}

func NewDatabase() (Database, error) {
	sourceUrl := "https://raw.githubusercontent.com/vikpe/qw-servers-geoip/main/ip_to_geo.json"
	destPath := "ip_to_geo.json"
	err := downloadFile(sourceUrl, destPath)
	if err != nil {
		return nil, err
	}

	geoJsonFile, _ := os.ReadFile(destPath)

	var geoDatabase Database
	err = json.Unmarshal(geoJsonFile, &geoDatabase)
	if err != nil {
		return nil, err
	}

	return geoDatabase, nil
}

func downloadFile(url string, dest string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}
