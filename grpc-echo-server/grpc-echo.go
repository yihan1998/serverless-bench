package main

import (
	"fmt"
	"net"

	ping "github.com/yihan1998/serverless-bench/grpc-echo-go/proto"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var port = 8080

type pingServer struct {
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