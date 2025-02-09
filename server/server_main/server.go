package main

import (
	"context"
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
	"google.golang.org/grpc/peer"
)

var addr, certFn, keyFn string

func init() {
	flag.StringVar(&addr, "address", "localhost:34443", "listen address")
	flag.StringVar(&certFn, "cert", "cert.pem", "certificate file")
	flag.StringVar(&keyFn, "key", "key.pem", "private key file")
}

// UnaryInterceptor는 단일 RPC 호출을 가로채는 인터셉터입니다.
func UnaryInterceptor(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (interface{}, error) {
	// 클라이언트 정보 로깅
	if p, ok := peer.FromContext(ctx); ok {
		fmt.Printf("Client connected: %s\n", p.Addr)
	}
	// 실제 핸들러 호출
	resp, err := handler(ctx, req)

	// 클라이언트 연결 해제 로깅
	if p, ok := peer.FromContext(ctx); ok {
		fmt.Printf("client disconnected: %s\n", p.Addr)
	}

	return resp, err
}

// wrappedServerStream은 grpc.ServerStream을 래핑하여 스트림 종료 시 클라이언트 정보를 로깅합니다.
type wrappedServerStream struct {
	grpc.ServerStream
}

func (w *wrappedServerStream) RecvMsg(m interface{}) error {
	err := w.ServerStream.RecvMsg(m)
	if err == nil {
		return nil
	}
	if p, ok := peer.FromContext(w.Context()); ok {
		fmt.Printf("Client disconnected: %s\n", p.Addr)
	}
	return err
}

func (w *wrappedServerStream) SendMsg(m interface{}) error {
	err := w.ServerStream.SendMsg(m)
	if err == nil {
		return nil
	}
	if p, ok := peer.FromContext(w.Context()); ok {
		fmt.Printf("Client disconnected: %s\n", p.Addr)
	}
	return err
}

// StreamInterceptor는 스트림 RPC 호출을 가로채는 인터셉터입니다.
func StreamInterceptor(
	srv interface{},
	ss grpc.ServerStream,
	info *grpc.StreamServerInfo,
	handler grpc.StreamHandler,
) error {
	// 클라이언트 정보 로깅
	if p, ok := peer.FromContext(ss.Context()); ok {
		fmt.Printf("Client connected: %s\n", p.Addr)
	}
	// 실제 핸들러 호출
	err := handler(srv, &wrappedServerStream{ServerStream: ss})
	return err
}

func main() {
	flag.Parse()

	server := grpc.NewServer(
		grpc.UnaryInterceptor(UnaryInterceptor),   // 단일 RPC 인터셉터 추가
		grpc.StreamInterceptor(StreamInterceptor), // 스트림 RPC 인터셉터 추가
	)
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
		fmt.Println()
		signal := <-sig
		fmt.Printf("\nrecieve signal: %s, shutdown server... \n", signal)
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
