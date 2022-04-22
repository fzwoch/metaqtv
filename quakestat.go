package main

import (
	"strconv"
	"strings"
)

func parseClientRecord(clientRecord []string) (Client, error) {
	nameRawStr := clientRecord[ColName]

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
	team := quakeTextToPlainText(clientRecord[ColTeam])
	teamInt := stringToIntArray(team)
	frags, _ := strconv.Atoi(clientRecord[ColFrags])
	time_, _ := strconv.Atoi(clientRecord[ColTime])
	ping, _ := strconv.Atoi(clientRecord[ColPing])
	colorTop, _ := strconv.Atoi(clientRecord[ColColorTop])
	colorBottom, _ := strconv.Atoi(clientRecord[ColColorBottom])

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
