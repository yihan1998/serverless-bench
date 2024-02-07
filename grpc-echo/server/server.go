package main

import (
	"fmt"
	"net"

	echo "github.com/yihan1998/serverless-bench/grpc-echo/proto"
	"google.golang.org/grpc"
)

var port = 8080

type echoServer struct {
}

func main () {
	lis, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	pingServer := &pingServer{}

	grpcServer := grpc.NewServer()
	reflection.Register(grpcServer)
	ping.RegisterPingServiceServer(grpcServer, pingServer)
	grpcServer.Serve(lis)
}