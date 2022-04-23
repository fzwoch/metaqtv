package main

import (
	"net"
	"strconv"
	"sync"
)

type Player struct {
	Name    string
	NameInt []int
	Team    string
	TeamInt []int
	Skin    string
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
	Title         string
	Address       string
	Numspectators int
	Spectators    []string
}

type QuakeServer struct {
	Title         string
	Description   string
	Address       string
	Map           string
	NumPlayers    int
	MaxPlayers    int
	NumSpectators int
	MaxSpectators int
	Players       []Player
	Spectators    []Spectator
	Settings      map[string]string
	QtvAddress    string

	keepaliveCount int
}

func newQuakeServer() QuakeServer {
	return QuakeServer{
		Title:       "",
		Description: "",
		Address:     "",
		Settings:    map[string]string{},
		Players:     make([]Player, 0),
		Spectators:  make([]Spectator, 0),
		QtvAddress:  "",
	}
}

type RawServerSocketAddress struct {
	IpParts [4]byte
	Port    uint16
}

func (nsa RawServerSocketAddress) toSocketAddress() SocketAddress {
	ip := net.IPv4(nsa.IpParts[0], nsa.IpParts[1], nsa.IpParts[2], nsa.IpParts[3]).String()

	return SocketAddress{
		Host: ip,
		Port: int(nsa.Port),
	}
}

type SocketAddress struct {
	Host string
	Port int
}

func (sa SocketAddress) toString() string {
	return sa.Host + ":" + strconv.Itoa(sa.Port)
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
