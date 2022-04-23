package main

import (
	"bytes"
	"encoding/binary"
	"errors"
	"net"
)

var (
	requestStatusSequence = []byte{0x63, 0x0a, 0x00}
	validResponseSequence = []byte{0xff, 0xff, 0xff, 0xff, 0x64, 0x0a}
)

func ReadMasterServer(socketAddress string, retryCount int, timeout int) ([]SocketAddress, error) {
	addresses := make([]SocketAddress, 0)

	conn, err := net.Dial("udp4", socketAddress)
	if err != nil {
		return addresses, err
	}

	defer conn.Close()

	buffer := make([]byte, bufferMaxSize)
	bufferLength := 0

	for i := 0; i < retryCount; i++ {
		conn.SetDeadline(timeInFuture(timeout))

		_, err = conn.Write(requestStatusSequence)
		if err != nil {
			return addresses, err
		}

		conn.SetDeadline(timeInFuture(timeout))
		bufferLength, err = conn.Read(buffer)
		if err != nil {
			continue
		}

		break
	}

	if err != nil {
		return addresses, err
	}

	responseSequence := buffer[:len(validResponseSequence)]
	isValidSequence := bytes.Equal(responseSequence, validResponseSequence)

	if !isValidSequence {
		err = errors.New(socketAddress + ": Response error")
		return addresses, err
	}

	reader := bytes.NewReader(buffer[6:bufferLength])

	for {
		var rawAddress RawServerSocketAddress

		err = binary.Read(reader, binary.BigEndian, &rawAddress)
		if err != nil {
			break
		}

		addresses = append(addresses, rawAddress.toSocketAddress())
	}

	return addresses, nil
}
