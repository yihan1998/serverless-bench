package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"time"

	echo "github.com/yihan1998/serverless-bench/grpc-echo/echo"
	"google.golang.org/grpc"
)

var (
	serverAddr         = flag.String("server_addr", "127.0.0.1:8080", "The server address in the format of host:port")
	serverHostOverride = flag.String("server_host_override", "", "")
	insecure           = flag.Bool("insecure", false, "Set to true to skip SSL validation")
	skipVerify         = flag.Bool("skip_verify", false, "Set to true to skip server hostname verification in SSL validation")
)

var port = 8080

func main() {
	flag.Parse()

	opts := []grpc.DialOption{
		grpc.WithInsecure(),
	}

	conn, err := grpc.Dial(*serverAddr, opts...)
	if err != nil {
		log.Fatalf("fail to dial: %v", err)
	}
	defer conn.Close()

	client := echo.NewEchoServiceClient(conn)

	Echo(client, "hello")
	EchoStream(client, "hello")
}

func Echo(client echo.EchoServiceClient, msg string) {
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()
	rep, err := client.Echo(ctx, &echo.Request{Msg: msg})
	if err != nil {
		log.Fatalf("%v.Echo failed %v: ", client, err)
	}
	log.Printf("Echo got %v\n", rep.GetMsg())
}

func EchoStream(client echo.EchoServiceClient, msg string) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	stream, err := client.EchoStream(ctx)
	if err != nil {
		log.Fatalf("%v.(_) = _, %v", client, err)
	}

	waitc := make(chan struct{})
	go func() {
		for {
			in, err := stream.Recv()
			if err == io.EOF {
				// read done.
				close(waitc)
				return
			}
			if err != nil {
				log.Fatalf("Failed to receive a response : %v", err)
			}
			log.Printf("Got %s", in.GetMsg())
		}
	}()

	i := 0
	for i < 1 {
		if err := stream.Send(&echo.Request{Msg: fmt.Sprintf("%s-%d", msg, i)}); err != nil {
			log.Fatalf("Failed to send a ping: %v", err)
		}
		i++
	}
	stream.CloseSend()
	<-waitc
}
