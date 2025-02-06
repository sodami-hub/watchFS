package main

import (
	"flag"
	"fmt"
	"net"

	api "github.com/sodami-hub/watchfs/api/v1"
	"google.golang.org/grpc"
)

var addr string

func init() {
	flag.StringVar(&addr, "address", "localhost:34443", "listen address")
}

func main() {
	flag.Parse()

	server := grpc.NewServer()
	garageService := &GarageService{}
	api.RegisterGarageServer(server, garageService)

	listen, err := net.Listen("tcp", addr)
	if err != nil {
		fmt.Println(err)
		return
	}
	server.Serve(listen)
	fmt.Printf("Listening for TLS connections on %s ...", addr)
}
