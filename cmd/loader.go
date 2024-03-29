package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"golang.org/x/exp/slices"

	"github.com/vhive-serverless/loader/pkg/common"
	"github.com/vhive-serverless/loader/pkg/config"
	"github.com/vhive-serverless/loader/pkg/trace"

	"github.com/yihan1998/serverless-bench/pkg/driver"

	log "github.com/sirupsen/logrus"
	tracer "github.com/vhive-serverless/vSwarm/utils/tracing/go"
)

const (
	zipkinAddr = "http://localhost:9411/api/v2/spans"
)

var (
	configPath    = flag.String("config", "config.json", "Path to loader configuration file")
	verbosity     = flag.String("verbosity", "info", "Logging verbosity - choose from [info, debug, trace]")
	iatGeneration = flag.Bool("iatGeneration", false, "Generate iats only or run invocations as well")
	rate          = flag.Float64("rate", 1.0, "KReq per second")
	numWorker     = flag.Int("worker", 1, "Number of workers to generate workload")
)

func init() {
	flag.Parse()

	log.SetFormatter(&log.TextFormatter{
		TimestampFormat: time.StampMilli,
		FullTimestamp:   true,
	})
	log.SetOutput(os.Stdout)

	switch *verbosity {
	case "debug":
		log.SetLevel(log.DebugLevel)
	case "trace":
		log.SetLevel(log.TraceLevel)
	default:
		log.SetLevel(log.InfoLevel)
	}
}

func main() {
	cfg := config.ReadConfigurationFile(*configPath)

	if cfg.EnableZipkinTracing {
		// TODO: how not to exclude Zipkin spans here? - file a feature request
		log.Warnf("Zipkin tracing has been enabled. This will exclude Istio spans from the Zipkin traces.")

		shutdown, err := tracer.InitBasicTracer(zipkinAddr, "loader")
		if err != nil {
			log.Print(err)
		}

		defer shutdown()
	}

	if cfg.ExperimentDuration < 1 {
		log.Fatal("Runtime duration should be longer, at least a minute.")
	}

	supportedPlatforms := []string{
		"Knative",
	}

	if !slices.Contains(supportedPlatforms, cfg.Platform) {
		log.Fatal("Unsupported platform! Supported platforms are [Knative, OpenWhisk, AWSLambda, Dirigent]")
	}

	log.Infof("Running experiments for %d minutes on %s\n", cfg.ExperimentDuration, cfg.Platform)

	runTraceMode(&cfg, *iatGeneration)
}

func determineDurationToParse(runtimeDuration int, warmupDuration int) int {
	result := 0

	if warmupDuration > 0 {
		result += 1              // profiling
		result += warmupDuration // warmup
	}

	result += runtimeDuration // actual experiment

	return result
}

func runTraceMode(cfg *config.LoaderConfiguration, iatOnly bool) {
	durationToParse := determineDurationToParse(cfg.ExperimentDuration, cfg.WarmupDuration)

	traceParser := trace.NewAzureParser(cfg.TracePath, durationToParse)
	functions := traceParser.Parse(cfg.Platform)

	log.Infof("Traces contain the following %d functions:\n", len(functions))
	for _, function := range functions {
		fmt.Printf("\t%s\n", function.Name)
	}

	var iatType common.IatDistribution
	shiftIAT := false
	switch cfg.IATDistribution {
	case "exponential":
		iatType = common.Exponential
	case "exponential_shift":
		iatType = common.Exponential
		shiftIAT = true
	case "uniform":
		iatType = common.Uniform
	case "uniform_shift":
		iatType = common.Uniform
		shiftIAT = true
	case "equidistant":
		iatType = common.Equidistant
	default:
		log.Fatal("Unsupported IAT distribution.")
	}

	var yamlSpecificationPath string
	switch cfg.YAMLSelector {
	case "container":
		yamlSpecificationPath = "workloads/container/synthetic.yaml"
	default:
		if cfg.Platform != "Dirigent" {
			log.Fatal("Invalid 'YAMLSelector' parameter.")
		}
	}

	var traceGranularity common.TraceGranularity
	switch cfg.Granularity {
	case "minute":
		traceGranularity = common.MinuteGranularity
	case "second":
		traceGranularity = common.SecondGranularity
	default:
		log.Fatal("Invalid trace granularity parameter.")
	}

	log.Infof("Using %s as a service YAML specification file.\n", yamlSpecificationPath)

	experimentDriver := driver.NewDriver(&driver.DriverConfiguration{
		LoaderConfiguration: cfg,
		IATDistribution:     iatType,
		ShiftIAT:            shiftIAT,
		TraceGranularity:    traceGranularity,
		TraceDuration:       durationToParse,

		Rate:      rate,
		NumWorker: numWorker,

		YAMLPath: yamlSpecificationPath,
		TestMode: false,

		Functions: functions,
	})

	experimentDriver.RunExperiment(iatOnly)
}
