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

	proxied, err := DialTcp(*proxyHost, *proxyPort)
	if err != nil {
		panic(err)
	}
	defer proxied.Close()

	// though it seems like it can handle more, it actually supports only one connection,
	err = Listen("localhost", *port, NewProxyHandler(proxied))
	if err != nil {
		panic(err)
	}

}
