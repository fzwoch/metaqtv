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

	port, _ := strconv.Atoi(strings.Split(mvdsv.Address, ":")[1])

	return GameState{
		Hostname:  mvdsv.Address,
		IpAddress: mvdsv.Address,
		Port:      port,
		Link:      fmt.Sprintf("http://%s/watch.qtv?sid=3", mvdsv.Address),
		Players:   players,
	}
}

func ToGameStates(servers []mvdsv.MvdsvExport) any {
	states := make([]GameState, 0)

	for _, s := range servers {
		if s.QtvStream.Url != "" {
			states = append(states, GameStateFromMdvsv(s))
		}
	}

	return states
}

func ServersHandler(serverSource func() []mvdsv.MvdsvExport) http.HandlerFunc {
	getGameStates := func() any { return ToGameStates(serverSource()) }
	return mhttp.CreateHandler(getGameStates)
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
