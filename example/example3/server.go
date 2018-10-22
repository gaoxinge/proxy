package main

import "proxy"

func main() {
	socks4Proxy := proxy.NewSocks4Proxy("localhost:8080")
	socks4Proxy.Start("")
}
