package main

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

func isBotName(name string) bool {
	switch name {
	case
		"[ServeMe]",
		"twitch.tv/vikpe":
		return true
	}
	return false
}

func isBotPing(ping int) bool {
	switch ping {
	case
		10:
		return true
	}
	return false
}

func parseClientRecord(clientRecord []string) (Client, error) {
	const (
		IndexFrags              = 1
		IndexTime               = 2
		IndexPing               = 3
		IndexName               = 4
		IndexColorTop           = 6
		IndexColorBottom        = 7
		IndexTeam               = 8
		SpectatorPrefix  string = "\\s\\"
	)

	expectedColumnCount := 9
	columnCount := len(clientRecord)

	if columnCount != expectedColumnCount {
		err := errors.New(fmt.Sprintf("invalid player column count %d.", columnCount))
		return Client{}, err
	}

	nameRawStr := clientRecord[IndexName]

	isSpec := strings.HasPrefix(nameRawStr, SpectatorPrefix)
	if isSpec {
		nameRawStr = strings.TrimPrefix(nameRawStr, SpectatorPrefix)
	}

	name := quakeTextToPlainText(nameRawStr)
	nameInt := stringToIntArray(name)
	team := quakeTextToPlainText(clientRecord[IndexTeam])
	teamInt := stringToIntArray(team)
	frags, _ := strconv.Atoi(clientRecord[IndexFrags])
	time_, _ := strconv.Atoi(clientRecord[IndexTime])
	ping, _ := strconv.Atoi(clientRecord[IndexPing])
	colorTop, _ := strconv.Atoi(clientRecord[IndexColorTop])
	colorBottom, _ := strconv.Atoi(clientRecord[IndexColorBottom])

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
			IsBot:   isBotName(name) || isBotPing(ping),
		},
		IsSpec: isSpec,
	}, nil

}
