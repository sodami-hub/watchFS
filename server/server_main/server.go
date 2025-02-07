package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	api "github.com/sodami-hub/watchfs/api/v1"
	service "github.com/sodami-hub/watchfs/server"
	"google.golang.org/grpc"
)

var addr, certFn, keyFn string

func init() {
	flag.StringVar(&addr, "address", "localhost:34443", "listen address")
	flag.StringVar(&certFn, "cert", "cert.pem", "certificate file")
	flag.StringVar(&keyFn, "key", "key.pem", "private key file")
}

func main() {
	flag.Parse()

	server := grpc.NewServer()
	garageService := &service.GarageService{}
	api.RegisterGarageServer(server, garageService)

	cert, err := tls.LoadX509KeyPair(certFn, keyFn)
	if err != nil {
		fmt.Println(err)
		return
	}

	listen, err := net.Listen("tcp", addr)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("Listening for TLS connections on %s ...", addr)

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		signal := <-sig
		fmt.Printf("recieve signal: %s, shutdown server... \n", signal)
		server.GracefulStop()
	}()

	log.Fatal(server.Serve(tls.NewListener(listen,
		&tls.Config{
			Certificates:             []tls.Certificate{cert},
			CurvePreferences:         []tls.CurveID{tls.CurveP256},
			MinVersion:               tls.VersionTLS12,
			PreferServerCipherSuites: true,
			NextProtos:               []string{"h2"},
			// ALPN(Application-Layer Protocol Negotiation) 속성 설정 h2는 서버와 클라이언트가 모두 HTTP/2를 지원하는 경우 h2 프로토콜을 사용해서 통신한다.
		},
	)))

}
