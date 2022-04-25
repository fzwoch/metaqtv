package main

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
)

type GeoInfo struct {
	Alpha2  string
	Country string
	Region  string
}

func getGeoData() (map[string]GeoInfo, error) {
	sourceUrl := "https://raw.githubusercontent.com/vikpe/qw-servers-geoip/main/ip_to_geo.json"
	destPath := "ip_to_geo.json"
	err := downloadFile(sourceUrl, destPath)
	if err != nil {
		return nil, err
	}

	geoJsonFile, _ := os.ReadFile(destPath)

	var geoData map[string]GeoInfo
	err = json.Unmarshal(geoJsonFile, &geoData)
	if err != nil {
		return nil, err
	}

	return geoData, nil
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
