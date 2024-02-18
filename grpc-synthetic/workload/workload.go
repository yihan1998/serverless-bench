package workload

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	log "github.com/sirupsen/logrus"
	util "github.com/vhive-serverless/loader/pkg/common"
	"github.com/yihan1998/serverless-bench/grpc-synthetic/proto"
	tracing "github.com/vhive-serverless/vSwarm/utils/tracing/go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// static double SQRTSD (double x) {
//     double r;
//     __asm__ ("sqrtsd %1, %0" : "=x" (r) : "x" (x));
//     return r;
// }
import "C"

const (
	// ContainerImageSizeMB was chosen as a median of the container physical memory usage.
	// Allocate this much less memory inside the actual function.
	ContainerImageSizeMB = 15
)

const EXEC_UNIT int = 1e2

var hostname string
var IterationsMultiplier int

func takeSqrts() C.double {
	var tmp C.double // Circumvent compiler optimizations
	for i := 0; i < EXEC_UNIT; i++ {
		tmp = C.SQRTSD(C.double(10))
	}
	return tmp
}

type funcServer struct {
	proto.UnimplementedExecutorServer
}

func busySpin(runtimeMilli uint32) {
	totalIterations := IterationsMultiplier * int(runtimeMilli)

	for i := 0; i < totalIterations; i++ {
		takeSqrts()
	}
}

func TraceFunctionExecution(start time.Time, timeLeftMicroseconds uint32) (msg string) {
	timeConsumedMicroseconds := uint32(time.Since(start).Microseconds())
	if timeConsumedMicroseconds < timeLeftMicroseconds {
		timeLeftMicroseconds -= timeConsumedMicroseconds
		if timeLeftMicroseconds > 0 {
			busySpin(timeLeftMicroseconds)
		}

		msg = fmt.Sprintf("OK - %s", hostname)
	}

	return msg
}

func (s *funcServer) Execute(_ context.Context, req *proto.SynRequest) (*proto.SynReply, error) {
	var msg string
	start := time.Now()

	timeLeftMicroseconds := req.RuntimeInMicroSec
	msg = TraceFunctionExecution(start, timeLeftMicroseconds)

	return &proto.SynReply{
		Message:            msg,
		DurationInMicroSec: uint32(time.Since(start).Microseconds()),
		MemoryUsageInKb:    req.MemoryInMebiBytes * 1024,
	}, nil
}

func readEnvironmentalVariables() {
	if _, ok := os.LookupEnv("ITERATIONS_MULTIPLIER"); ok {
		IterationsMultiplier, _ = strconv.Atoi(os.Getenv("ITERATIONS_MULTIPLIER"))
	} else {
		// Cloudlab xl170 benchmark @ 1 second function execution time
		IterationsMultiplier = 102
	}

	log.Infof("ITERATIONS_MULTIPLIER = %d\n", IterationsMultiplier)

	var err error
	hostname, err = os.Hostname()
	if err != nil {
		log.Warn("Failed to get HOSTNAME environmental variable.")
		hostname = "Unknown host"
	}
}

func StartGRPCServer(serverAddress string, serverPort int, zipkinUrl string) {
	readEnvironmentalVariables()

	if tracing.IsTracingEnabled() {
		log.Infof("Zipkin URL: %s\n", zipkinUrl)
		shutdown, err := tracing.InitBasicTracer(zipkinUrl, "")
		if err != nil {
			log.Warn(err)
		}
		defer shutdown()
	}

	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", serverAddress, serverPort))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	var grpcServer *grpc.Server
	if tracing.IsTracingEnabled() {
		grpcServer = tracing.GetGRPCServerWithUnaryInterceptor()
	} else {
		grpcServer = grpc.NewServer()
	}

	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, syscall.SIGTERM)

	go func() {
		<-sigc
		log.Info("Received SIGTERM, shutting down gracefully...")
		grpcServer.GracefulStop()
	}()

	reflection.Register(grpcServer) // gRPC Server Reflection is used by gRPC CLI
	proto.RegisterExecutorServer(grpcServer, &funcServer{})
	err = grpcServer.Serve(lis)
	util.Check(err)
}
