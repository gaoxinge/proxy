package main

import "proxy"

func main() {
	socks5Proxy := proxy.NewSocks5Proxy("localhost:5050")
	socks5Proxy.Start("")
}
