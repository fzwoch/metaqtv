package main

import (
	"sync"
)

type Player struct {
	Name    string
	NameRaw []int
	Team    string
	TeamRaw []int
	Colors  [2]int
	Frags   int
	Ping    int
	Spec    bool
	Time    int
	IsBot   bool
}

type Spectator struct {
	Name    string
	NameRaw []int
	IsBot   bool
}

type QTV struct {
	Host          string
	Address       string
	Numspectators int
	Spectators    []string
}

type Server struct {
	IPAddress     string `json:"IpAddress"`
	Address       string
	Description   string
	Title         string
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
	QTV           []QTV

	keepaliveCount int
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
