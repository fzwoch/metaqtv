package v1

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/vikpe/serverstat/qserver/mvdsv"
	"metaqtv/dataprovider"
	"metaqtv/mhttp"
)

type Player struct {
	Name string
}

type GameState struct {
	Hostname  string
	IpAddress string
	Port      int
	Link      string
	Players   []Player
}

func GameStateFromMdvsv(mvdsv mvdsv.MvdsvExport) GameState {
	players := make([]Player, 0)

	for _, p := range mvdsv.Players {
		players = append(players, Player{Name: p.Name.ToPlainString()})
	}

	addressParts := strings.Split(mvdsv.Address, ":")
	ip := addressParts[0]
	port, _ := strconv.Atoi(addressParts[1])

	return GameState{
		Hostname:  ip,
		IpAddress: ip,
		Port:      port,
		Link:      fmt.Sprintf("http://%s/watch.qtv?sid=3", mvdsv.Address),
		Players:   players,
	}
}

func ToGameStates(servers []mvdsv.MvdsvExport) []GameState {
	states := make([]GameState, 0)

	for _, s := range servers {
		if s.QtvStream.Url != "" {
			states = append(states, GameStateFromMdvsv(s))
		}
	}

	return states
}

func ServersHandler(serverSource func() []mvdsv.MvdsvExport) http.HandlerFunc {
	getOutput := func() any {
		type serverStats struct {
			ServerCount       int
			ActiveServerCount int
			PlayerCount       int
			ObserverCount     int
		}

		servers := serverSource()

		stats := serverStats{
			ServerCount:       len(servers),
			ActiveServerCount: 0,
			PlayerCount:       0,
			ObserverCount:     0,
		}

		for _, s := range servers {
			if s.PlayerSlots.Used > 0 {
				stats.ActiveServerCount++
			}
			stats.PlayerCount += s.PlayerSlots.Used
			stats.ObserverCount += s.SpectatorSlots.Used
		}

		gameStates := ToGameStates(servers)

		type server struct{ GameStates []GameState }
		type result struct {
			Servers []server
			serverStats
		}

		return result{
			Servers: []server{
				{GameStates: gameStates},
			},
			serverStats: stats,
		}
	}
	return mhttp.CreateHandler(getOutput)
}

func New(baseUrl string, provider *dataprovider.DataProvider) mhttp.Api {
	return mhttp.Api{
		Provider: provider,
		BaseUrl:  baseUrl,
		Endpoints: mhttp.Endpoints{
			"servers": ServersHandler(provider.Mvdsv),
		},
	}
}
