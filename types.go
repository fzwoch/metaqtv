package main

import (
	"sync"
)

type Player struct {
	Name    string
	NameInt []int
	Team    string
	TeamInt []int
	Colors  [2]int
	Frags   int
	Ping    int
	Time    int
	IsBot   bool
}

type Client struct {
	Player
	IsSpec bool
}

type Spectator struct {
	Name    string
	NameInt []int
	IsBot   bool
}

type QtvServer struct {
	Host          string
	Address       string
	Numspectators int
	Spectators    []string
}

type QuakeServer struct {
	Title         string
	Description   string
	Ip            string `json:"IpAddress"`
	SocketAddress string
	Port          uint16
	Map           string
	NumClients    int
	NumPlayers    int
	MaxPlayers    int
	NumSpectators int
	MaxSpectators int
	Players       []Player
	Spectators    []Spectator
	Settings      map[string]string
	QTV           []QtvServer

	keepaliveCount int
}

func newQuakeServer() QuakeServer {
	return QuakeServer{
		Title:         "",
		Description:   "",
		Ip:            "",
		SocketAddress: "",
		Port:          0,
		Settings:      map[string]string{},
		Players:       make([]Player, 0),
		Spectators:    make([]Spectator, 0),
		QTV:           make([]QtvServer, 0),
	}
}

type SocketAddress struct {
	Ip   [4]byte
	Port uint16
}
type MasterServer struct {
	Hostname string `json:"hostname"`
	Port     int    `json:"port"`
}

type MutexStore struct {
	sync.RWMutex
	data []byte
}

func (store *MutexStore) Write(data []byte) {
	store.Lock()
	store.data = data
	store.Unlock()
}

func (store *MutexStore) Read() []byte {
	store.RLock()
	data := store.data
	store.RUnlock()
	return data
}

func newMutexStore() *MutexStore {
	store := MutexStore{data: make([]byte, 0)}
	return &store
}
