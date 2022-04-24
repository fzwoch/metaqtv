package main

import (
	"encoding/json"
	"os"
)

type geoData struct {
	Alpha2  string
	Country string
	Region  string
}

func getGeoData() map[string]geoData {
	sourceUrl := "https://raw.githubusercontent.com/vikpe/qw-servers-geoip/main/ip_to_geo.json"
	destPath := "ip_to_geo.json"
	Download(sourceUrl, destPath)

	geoJsonFile, _ := os.ReadFile(destPath)

	var geoData map[string]geoData
	json.Unmarshal(geoJsonFile, &geoData)

	return geoData
}

func newGeoData() geoData {
	return geoData{
		Alpha2:  "",
		Country: "",
		Region:  "",
	}
}
