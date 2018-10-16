package proxy

import (
	"net/http"
	"fmt"
	"net"
	"io"
	"log"
)

type HttpProxy struct{
	addr string
}

func NewHttpProxy(addr string) *HttpProxy {
	return &HttpProxy{addr}
}

func (httpProxy *HttpProxy) Start() {
	connect := NewConnect()
	http.ListenAndServe(httpProxy.addr, connect)
}

type Connect struct {}

func NewConnect() *Connect {
	return &Connect{}
}

func handleConnect(w http.ResponseWriter, r *http.Request) {
	server, err := net.Dial("tcp", r.Host)
	if err != nil {
		log.Println(err)
		return
	}

	hij, _ := w.(http.Hijacker)
	client, _, _ := hij.Hijack()
	client.Write([]byte("HTTP/1.0 200 Connection Established\r\n\r\n"))

	go io.Copy(server, client)
	io.Copy(client, server)
}

func (connect *Connect) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("Received request %s %s %s\n", r.Method, r.Host, r.RemoteAddr)

	if r.Method == "CONNECT" {
		handleConnect(w, r)
	} else {
		for k, v := range r.Header {
			fmt.Println(k, v)
		}
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte("This is a http tunnel proxy, only CONNECTED method is allowed"))
		return
	}
}
