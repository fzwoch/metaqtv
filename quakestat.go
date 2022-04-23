package main

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

const (
	ColIndexFrags              = 1
	ColIndexTime               = 2
	ColIndexPing               = 3
	ColIndexName               = 4
	ColIndexColorTop           = 6
	ColIndexColorBottom        = 7
	ColIndexTeam               = 8
	SpectatorPrefix     string = "\\s\\"
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
	expectedColumnCount := 9
	columnCount := len(clientRecord)

	if columnCount != expectedColumnCount {
		err := errors.New(fmt.Sprintf("invalid player column count %d.", columnCount))
		return Client{}, err
	}

	nameRawStr := clientRecord[ColIndexName]

	isSpec := strings.HasPrefix(nameRawStr, SpectatorPrefix)
	if isSpec {
		nameRawStr = strings.TrimPrefix(nameRawStr, SpectatorPrefix)
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
			IsBot:   isBotName(name) || isBotPing(ping),
		},
		IsSpec: isSpec,
	}, nil

}
