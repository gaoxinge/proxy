package proxy

import (
	"net/http"
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

type Connect struct {
}

func NewConnect() *Connect {
	return &Connect{}
}

func (connect *Connect) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Printf("receive %s request from %s to %s\n", r.Method, r.RemoteAddr, r.Host)

	if r.Method == "CONNECT" {
		handleHttpConnect(w, r)
	} else {
		handleHttpMethod(w, r)
	}
}

func handleHttpConnect(w http.ResponseWriter, r *http.Request) {
	server, err := net.Dial("tcp", r.Host)
	if err != nil {
		log.Println(err)
		return
	}

	hij, _ := w.(http.Hijacker)
	client, _, _ := hij.Hijack()
	client.Write([]byte("HTTP/1.0 200 Connection Established\r\n\r\n"))


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

func handleHttpMethod(w http.ResponseWriter, r *http.Request) {
	req, err := copyRequest(r)
	if err != nil {
		log.Println(err)
		return
	}

	var server http.RoundTripper
	server = http.DefaultTransport

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
	if _, err := io.Copy(w, resp.Body); err != nil {
		log.Println(err)
	}
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
