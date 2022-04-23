package main

import (
	"strconv"
	"sync"
)

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
