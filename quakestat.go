package main

import (
	"strconv"
	"strings"
)

const (
	ColIndexFrags       = 1
	ColIndexTime        = 2
	ColIndexPing        = 3
	ColIndexName        = 4
	ColIndexColorTop    = 6
	ColIndexColorBottom = 7
	ColIndexTeam        = 8
)

func parseClientRecord(clientRecord []string) (Client, error) {
	nameRawStr := clientRecord[ColIndexName]

	isBot := false

	if strings.HasSuffix(nameRawStr, "[ServeMe]") {
		isBot = true
	}

	isSpec := false
	spectatorPrefix := "\\s\\"
	if strings.HasPrefix(nameRawStr, spectatorPrefix) {
		nameRawStr = strings.TrimPrefix(nameRawStr, spectatorPrefix)
		isSpec = true
	}

	name := quakeTextToPlainText(nameRawStr)
	nameInt := stringToIntArray(name)
	team := quakeTextToPlainText(clientRecord[ColIndexTeam])
	teamInt := stringToIntArray(team)
	frags, _ := strconv.Atoi(clientRecord[ColIndexFrags])
	time_, _ := strconv.Atoi(clientRecord[ColIndexTime])
	ping, _ := strconv.Atoi(clientRecord[ColIndexPing])
	colorTop, _ := strconv.Atoi(clientRecord[ColIndexColorTop])
	colorBottom, _ := strconv.Atoi(clientRecord[ColIndexColorBottom])

	return Client{
		Player: Player{
			Name:    name,
			NameInt: nameInt,
			Team:    team,
			TeamInt: teamInt,
			Colors:  [2]int{colorTop, colorBottom},
			Frags:   frags,
			Ping:    ping,
			Time:    time_,
			IsBot:   isBot,
		},
		IsSpec: isSpec,
	}, nil

}
