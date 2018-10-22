package main

import "proxy"

func main() {
	httpProxy := proxy.NewHttpProxy("localhost:5050")
	httpProxy.Start("")
}
