package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/yihan1998/serverless-bench/grpc-synthetic/proto"
	"google.golang.org/grpc"
)

var (
	serverAddr         = flag.String("server_addr", "grpc-echo.default.10.200.3.4.sslip.io:80", "The server address in the format of host:port")
	serverHostOverride = flag.String("server_host_override", "", "")
	insecure           = flag.Bool("insecure", false, "Set to true to skip SSL validation")
	skipVerify         = flag.Bool("skip_verify", false, "Set to true to skip server hostname verification in SSL validation")
)

var port = 80

func main() {
	flag.Parse()

	opts := []grpc.DialOption{
		grpc.WithInsecure(),
	}

	dialContext, cancelDialing := context.WithTimeout(context.Background(), time.Duration(10)*time.Second)
	defer cancelDialing()

	conn, err := grpc.DialContext(dialContext, *serverAddr, opts...)
	if err != nil {
		log.Fatalf("fail to dial: %v", err)
	}
	defer conn.Close()

	client := proto.NewExecutorClient(conn)

	response, err := grpcClient.Execute(executionCxt, &proto.SynRequest{
		Message:           "nothing",
	})

	if err != nil {
		log.Debugf("gRPC timeout exceeded - %s", err)
	}
}
