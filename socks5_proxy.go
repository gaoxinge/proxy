package proxy

import (
	"net"
	"log"
	"bufio"
	"errors"
	"fmt"
	"encoding/binary"
	"io"
)

type Socks5Proxy struct {
	addr string
}

func NewSocks5Proxy(addr string) *Socks5Proxy {
	return &Socks5Proxy{addr}
}

func (socks5Proxy *Socks5Proxy) Start(nextAddr string) {
	broker, err := net.Listen("tcp", socks5Proxy.addr)
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

		go serveSocks5(client, nextAddr)
	}
}

func serveSocks5(client net.Conn, nextAddr string) {
	defer client.Close()

	if nextAddr == "" {
		log.Printf("client address: %s\n", client.RemoteAddr())
	} else {
		log.Printf("client address: %s, next address: %s\n", client.RemoteAddr(), nextAddr)
	}

	if nextAddr == "" {
		handShakeSocks5Fst(client)
		handShackSocks5Sec(client)
	} else {
		handShakeSocks5Other(client, nextAddr)
	}
}

func handShakeSocks5Fst(client net.Conn) {
	reader := bufio.NewReader(client)
	b, err := reader.Peek(3)
	if err != nil {
		log.Println(err)
		return
	}

	if b[0] != 0x05 {
		log.Println(errors.New(fmt.Sprintf("version number %b is not socks5 protocol", b[0])))
		return
	}

	switch b[2] {
	case 0x00:
		handleSocks5WOAuth(client)
	case 0x01, 0x02, 0x03, 0x80, 0xff:
		log.Println(errors.New(fmt.Sprintf("method %b is not supported by socks5proxy", b[2])))
	default:
		log.Println(errors.New(fmt.Sprintf("method %b is not supported by socks5 protocol", b[2])))
	}
}

func handShackSocks5Sec(client net.Conn) {
	b := make([]byte, 5)
	n, err := io.ReadAtLeast(client, b, 5)
	if err != nil {
		log.Println("qwer", err)
		return
	}

	if b[0] != 0x05 {
		log.Println(errors.New(fmt.Sprintf("version number %b is not socks5 protocol", b[0])))
		return
	}

	var length int
	switch b[3] {
	case 0x01:
		length = 10
	case 0x03:
		length = int(b[4]) + 7
	case 0x04:
		length = 22
	default:
		log.Println(errors.New(fmt.Sprintf("address type %b is supported by socks5 protocol", b[3])))
	}

	if n < length {
		b = append(b, make([]byte, length - n)...)
		_, err := io.ReadFull(client, b[n:length])
		if err != nil {
			log.Println("qwerqwer", err)
			return
		}
	}

	var host string
	switch b[3] {
	case 0x01:
		host = net.IP(b[4:8]).String()
	case 0x03:
		host = string(b[5:b[4]+5])
	case 0x04:
		host = net.IP(b[4:20]).String()
	}
	port := fmt.Sprintf("%d", binary.BigEndian.Uint16(b[len(b)-2:]))
	addr := net.JoinHostPort(host, port)

	switch b[1] {
	case 0x01:
		handleSocks5Connect(client, addr, b)
	case 0x02:
		log.Println(errors.New(fmt.Sprintf("command %b is not supproted by socks4proxy", b[1])))
	case 0x03:
		handleSocks5UDP(client, addr, b)
	default:
		log.Println(errors.New(fmt.Sprintf("command %b is not supported by socks4 protocol", b[1])))
	}
}

func handShakeSocks5Other(client net.Conn, nextAddr string) {
	handleSocks5Other(client, nextAddr)
}

func handleSocks5WOAuth(client net.Conn) {
	client.Write([]byte{0x05, 0x00})
}

func handleSocks5Connect(client net.Conn, addr string, b []byte) {
	server, err := net.Dial("tcp", addr)
	if err != nil {
		b[1] = 0x01
		client.Write(b)
		log.Println(err)
		return
	}
	b[1] = 0x00
	client.Write(b)

	transportTCP(client, server)
}

func handleSocks5UDP(client net.Conn, addr string, b []byte) {
	server, err := net.Dial("udp", addr)
	if err != nil {
		b[1] = 0x01
		client.Write(b)
		log.Println(err)
		return
	}
	b[1] = 0x00
	client.Write(b)

	transportUDP(client, server)
}

func handleSocks5Other(client net.Conn, nextAddr string) {
	server, err := net.Dial("tcp", nextAddr)
	if err != nil {
		log.Println(err)
		return
	}

	transportTCP(client, server)
}
