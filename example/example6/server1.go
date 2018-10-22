package main

import "proxy"

func main() {
	socks5Proxy := proxy.NewSocks5Proxy("localhost:8080")
	socks5Proxy.Start("localhost:5050")
}
