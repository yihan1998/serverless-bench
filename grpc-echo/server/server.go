package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"time"

	echo "github.com/yihan1998/serverless-bench/grpc-echo/proto"
	"google.golang.org/grpc"
)

var port = 8080

type echoServer struct {
}

func (e *echoServer) Echo(ctx context.Context, req *echo.Request) (*echo.Response, error) {
	log.Printf("Received: %s", req)
	return &echo.Response{Msg: fmt.Sprintf("%s - pong", req.Msg)}, nil
}

func (e *echoServer) EchoStream(stream echo.EchoService_EchoStreamServer) error {
	for {
		req, err := stream.Recv()

		if err == io.EOF {
			fmt.Println("Client disconnected")
			return nil
		}

		if err != nil {
			fmt.Println("Failed to receive ping")
			return err
		}

		fmt.Printf("Replying to ping %s at %s\n", req.Msg, time.Now())

		err = stream.Send(&ping.Response{
			Msg: fmt.Sprintf("pong %s", time.Now()),
		})

		if err != nil {
			fmt.Printf("Failed to send pong %s\n", err)
			return err
		}
	}
}

func main() {
	lis, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	echoServer := &echoServer{}

	grpcServer := grpc.NewServer()
	echo.RegisterEchoServiceServer(grpcServer, echoServer)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}

}
