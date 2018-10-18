package main

import "proxy"

func main() {
	httpProxy := proxy.NewHttpProxy("localhost:8080")
	httpProxy.Start("localhost:5050")
}