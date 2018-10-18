package main

import "proxy"

func main() {
	sock4Proxy := proxy.NewSocks4Proxy("localhost:8080")
	sock4Proxy.Start()
}
