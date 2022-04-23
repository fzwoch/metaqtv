package main

import (
	"strconv"
)

type SocketAddress struct {
	Host string
	Port int
}

func (sa SocketAddress) toString() string {
	return sa.Host + ":" + strconv.Itoa(sa.Port)
}
