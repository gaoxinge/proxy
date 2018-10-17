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

func (httpProxy *HttpProxy) Start(nextAddr string) {
	connect := NewConnect(nextAddr)
	http.ListenAndServe(httpProxy.addr, connect)
}

type Connect struct {
	nextAddr string
}

func NewConnect(nextAddr string) *Connect {
	return &Connect{nextAddr}
}

func (connect *Connect) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("Received request %s %s %s %s %s %s\n", r.Method, r.Host, r.RemoteAddr, r.RequestURI, r.URL, r.URL.String())

	if connect.nextAddr == "" {
		if r.Method == "CONNECT" {
			handleConnect(w, r)
		} else {
			handleMethod(w, r)
		}
	} else {
		if r.Method == "CONNECT" {

		} else {

		}
	}
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

func handleMethod(w http.ResponseWriter, r *http.Request) {
	req, err := copyRequest(r)
	if err != nil {
		log.Println(err)
		return
	}

	server := http.DefaultTransport
	resp, err := server.RoundTrip(req)
	if err != nil {
		log.Println(err)
		return
	}

	for key, value := range resp.Header {
		for _, v := range value {
			w.Header().Add(key, v)
		}
	}
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
	resp.Body.Close()
}

func copyRequest(r *http.Request) (*http.Request, error) {
	req, err := http.NewRequest(r.Method, r.URL.String(), r.Body)
	if err != nil {
		return nil, err
	}

	for key, value := range r.Header {
		for _, v := range value {
			req.Header.Add(key, v)
		}
	}

	if proxyConn := req.Header.Get("Proxy-Connection"); proxyConn != "" {
		req.Header.Del("Proxy-Connection")
		req.Header.Set("Connection", proxyConn)
	}

	if ip, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
		if xff := req.Header.Get("X-Forwarded-For"); xff != "" {
			ip = xff + ", " + ip
		}
		req.Header.Set("X-Forwarded-For", ip)
	}

	return req, nil
}
