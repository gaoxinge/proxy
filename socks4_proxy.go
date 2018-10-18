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
		}

		go serve(client)
	}
}

func serve(client net.Conn) {
	defer client.Close()

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

	switch b[1] {
	case 0x01:
		handleSocks4Connect(client, b)
	case 0x02:
		log.Println(errors.New(fmt.Sprintf("operation %b is not supproted by socks4proxy", b[1])))
	default:
		log.Println(errors.New(fmt.Sprintf("operation %b is not supported by socks4 protocol", b[1])))
	}
}

func handleSocks4Connect(client net.Conn, b []byte) {
	port := binary.BigEndian.Uint16(b[2:4])
	host := net.IP(b[4:8])
	addr := net.JoinHostPort(host.String(), fmt.Sprintf("%d", port))

	server, err := net.Dial("tcp", addr)
	if err != nil {
		client.Write([]byte{0x00, 0x5b, b[2], b[3], b[4], b[5], b[6], b[7]})
		log.Println(err)
		return
	}
	client.Write([]byte{0x00, 0x5a, b[2], b[3], b[4], b[5], b[6], b[7]})

	done := make(chan struct{})
	go func() {
		if _, err := io.Copy(server, client); err != nil {
			log.Println(err)
		}
		tcpServer := server.(*net.TCPConn)
		tcpServer.CloseWrite()
		done <- struct{}{}
	}()
	if _, err := io.Copy(client, server); err != nil {
		log.Println(err)
	}
	tcpServer := server.(*net.TCPConn)
	tcpServer.CloseRead()
	<- done
}