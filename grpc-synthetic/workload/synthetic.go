package main

import (
	"flag"
	"os"
	"strconv"

	log "github.com/sirupsen/logrus"
	"github.com/yihan1998/serverless-bench/grpc-synthetic/workload"
)

var (
	zipkin = flag.String("zipkin", "http://zipkin.zipkin:9411/api/v2/spans", "zipkin url")
)

func main() {
	var serverPort = 80

	if _, ok := os.LookupEnv("FUNC_PORT_ENV"); ok {
		serverPort, _ = strconv.Atoi(os.Getenv("FUNC_PORT_ENV"))
	}

	log.Infof("Port: %d\n", serverPort)
	workload.StartGRPCServer("", serverPort, *zipkin)
}