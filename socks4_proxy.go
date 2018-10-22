package proxy

import (
	"net"
	"log"
	"bufio"
	"errors"
	"fmt"
	"encoding/binary"
)

type Socks4Proxy struct {
	addr string
}

func NewSocks4Proxy(addr string) *Socks4Proxy {
	return &Socks4Proxy{addr}
}

func (socks4Proxy *Socks4Proxy) Start(nextAddr string) {
	broker, err := net.Listen("tcp", socks4Proxy.addr)
	if err != nil {
		log.Println(err)
		return
	}

	defer broker.Close()

	for {
		client, err := broker.Accept()
		if err != nil {
			log.Println(err)
			continue
		}

		go serveSocks4(client, nextAddr)
	}
}

func serveSocks4(client net.Conn, nextAddr string) {
	defer client.Close()

	if nextAddr == "" {
		log.Printf("client address: %s\n", client.RemoteAddr())
	} else {
		log.Printf("client address: %s, next address: %s\n", client.RemoteAddr(), nextAddr)
	}

	if nextAddr == "" {
		handShakeSocks4(client)
	} else {
		handShakeSocks4Other(client, nextAddr)
	}
}

func handShakeSocks4(client net.Conn) {
	reader := bufio.NewReader(client)
	b, err := reader.Peek(9)
	if err != nil {
		log.Println(err)
		return
	}

	if b[0] != 0x04 {
		log.Println(errors.New(fmt.Sprintf("version number %b is not socks4 protocol", b[0])))
		return
	}

	port := binary.BigEndian.Uint16(b[2:4])
	host := net.IP(b[4:8])
	addr := net.JoinHostPort(host.String(), fmt.Sprintf("%d", port))

	switch b[1] {
	case 0x01:
		handleSocks4Connect(client, addr, b)
	case 0x02:
		log.Println(errors.New(fmt.Sprintf("operation %b is not supproted by socks4proxy", b[1])))
	default:
		log.Println(errors.New(fmt.Sprintf("operation %b is not supported by socks4 protocol", b[1])))
	}
}

func handShakeSocks4Other(client net.Conn, nextAddr string) {
	handleSocks4Other(client, nextAddr)
}

func handleSocks4Connect(client net.Conn, addr string, b []byte) {
	server, err := net.Dial("tcp", addr)
	if err != nil {
		client.Write([]byte{0x00, 0x5b, b[2], b[3], b[4], b[5], b[6], b[7]})
		log.Println(err)
		return
	}
	client.Write([]byte{0x00, 0x5a, b[2], b[3], b[4], b[5], b[6], b[7]})

	transportTCP(client, server)
}

func handleSocks4Other(client net.Conn, nextAddr string) {
	server, err := net.Dial("tcp", nextAddr)
	if err != nil {
		log.Println(err)
		return
	}

	transportTCP(client, server)
}