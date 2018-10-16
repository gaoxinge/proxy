package proxy

import (
	"net/http"
	"fmt"
	"net"
	"io"
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

func (connect *Connect) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("Received request %s %s %s\n", r.Method, r.Host, r.RemoteAddr)

	if r.Method != "CONNECT" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte("This is a http tunnel proxy, only CONNECTED method is allowed"))
		return
	}

	host := r.URL.Host
	hij, ok := w.(http.Hijacker)
	if !ok {
		panic("HTTP Server does not support hijacking")
	}

	client, _, err := hij.Hijack()
	if err != nil {
		return
	}

	server, err := net.Dial("tcp", host)
	if err != nil {
		return
	}
	client.Write([]byte("HTTP/1.0 200 Connection Established\r\n\r\n"))

	go io.Copy(server, client)
	io.Copy(client, server)
}
