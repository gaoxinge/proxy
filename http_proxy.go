package proxy

import (
	"net/http"
	"net"
	"io"
	"log"
	"net/url"
	"time"
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

func (c *Connect) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if c.nextAddr == "" {
		log.Printf("receive %s request from %s to %s\n", r.Method, r.RemoteAddr, r.Host)
	} else {
		log.Printf("receive %s request from %s to %s\n", r.Method, r.RemoteAddr, c.nextAddr)
	}

	if r.Method == "CONNECT" {
		handleHttpConnect(w, r, c)
	} else {
		handleHttpMethod(w, r, c)
	}
}

func handleHttpConnect(w http.ResponseWriter, r *http.Request, c *Connect) {
	var server net.Conn
	var err error
	if c.nextAddr == "" {
		server, err = net.Dial("tcp", r.Host)
	} else {
		server, err = net.Dial("tcp", c.nextAddr)
	}
	if err != nil {
		log.Println(err)
		return
	}

	hij, _ := w.(http.Hijacker)
	client, _, _ := hij.Hijack()

	if c.nextAddr == "" {
		client.Write([]byte("HTTP/1.0 200 Connection Established\r\n\r\n"))
	} else {
		server.Write([]byte(r.Method + " " + r.RequestURI + " " + r.Proto + "\r\n\r\n"))
	}

	transport(client, server)
}

func handleHttpMethod(w http.ResponseWriter, r *http.Request, c *Connect) {
	req, err := copyRequest(r, c)
	if err != nil {
		log.Println(err)
		return
	}

	var server http.RoundTripper
	if c.nextAddr == "" {
		server = http.DefaultTransport
	} else {
		proxy := func(r *http.Request) (*url.URL, error) {
			return url.Parse("http://" + c.nextAddr)
		}
		server = &http.Transport{
			Proxy: proxy,
			DialContext: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
				DualStack: true,
			}).DialContext,
			MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		}
	}

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

func copyRequest(r *http.Request, c *Connect) (*http.Request, error) {
	req, err := http.NewRequest(r.Method, r.URL.String(), r.Body)
	if err != nil {
		return nil, err
	}

	for key, value := range r.Header {
		for _, v := range value {
			req.Header.Add(key, v)
		}
	}

	if c.nextAddr == "" {
		if proxyConn := req.Header.Get("Proxy-Connection"); proxyConn != "" {
			req.Header.Del("Proxy-Connection")
			req.Header.Set("Connection", proxyConn)
		}
	}

	if ip, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
		if xff := req.Header.Get("X-Forwarded-For"); xff != "" {
			ip = xff + ", " + ip
		}
		req.Header.Set("X-Forwarded-For", ip)
	}

	return req, nil
}
