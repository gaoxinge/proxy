package main

import "proxy"

func main() {
	socks4Proxy := proxy.NewSocks4Proxy("localhost:5050")
	socks4Proxy.Start("")
}