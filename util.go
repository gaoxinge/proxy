package proxy

import (
	"net"
	"io"
	"log"
)

func transportTCP(client net.Conn, server net.Conn) {
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

func transportUDP(client net.Conn, server net.Conn) {
	done := make(chan struct{})
	go func() {
		if _, err := io.Copy(server, client); err != nil {
			log.Println(err)
		}
		done <- struct{}{}
	}()
	if _, err := io.Copy(client, server); err != nil {
		log.Println(err)
	}
	server.Close()
	<- done
}