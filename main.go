package main

import (
	"flag"
	"fmt"
)

var debug = flag.Bool("debug", false, "enable debug log")
var port = flag.Int("port", 0, "host port")
var proxyHost = flag.String("proxy-host", "localhost", "proxied host")
var proxyPort = flag.Int("proxy-port", 80, "proxied port")

func main() {
	flag.Parse()
	if *port == 0 {
		fmt.Println("port is required")
		return
	}
	if *proxyHost == "" {
		fmt.Println("proxy host is required")
		return
	}
	if *proxyPort == 0 {
		fmt.Println("proxy port is required")
		return
	}
	fmt.Printf("tcp-proxy version: %v\n", Version)

	err := Listen("localhost", *port, NewProxyHandler(ProxyTarget{Host: *proxyHost, Port: *proxyPort}))
	if err != nil {
		panic(err)
	}

}
