package driver

import (
	"sync"
	"sync/atomic"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/vhive-serverless/loader/pkg/common"
	"github.com/vhive-serverless/loader/pkg/config"
	"github.com/vhive-serverless/loader/pkg/driver"
	"github.com/vhive-serverless/loader/pkg/generator"
	"github.com/vhive-serverless/loader/pkg/trace"

	"github.com/yihan1998/serverless-bench/pkg/distribution"
)

type DriverConfiguration struct {
	LoaderConfiguration *config.LoaderConfiguration
	IATDistribution     common.IatDistribution
	ShiftIAT            bool // shift the invocations inside minute
	TraceGranularity    common.TraceGranularity
	TraceDuration       int // in minutes

	YAMLPath string
	TestMode bool

	Functions []*common.Function
}

type Driver struct {
	Configuration          *DriverConfiguration
	SpecificationGenerator *generator.SpecificationGenerator
}

func NewDriver(driverConfig *DriverConfiguration) *Driver {
	return &Driver{
		Configuration:          driverConfig,
		SpecificationGenerator: generator.NewSpecificationGenerator(driverConfig.LoaderConfiguration.Seed),
	}
}

func (c *DriverConfiguration) WithWarmup() bool {
	if c.LoaderConfiguration.WarmupDuration > 0 {
		return true
	} else {
		return false
	}
}

func (d *Driver) invokeFunction(startTime time.Time, duration int) {
	var arrivalGenerator = dist.exponentialGenerator(1.0)
	var lastInvokeTime = time.Now()
	var nextInterval = dist.getNext(arrivalGenerator)

	log.Debug("Next interval: ", nextInterval)

	for {
		currentTime := time.Now()

		if int(currentTime.Sub(startTime).Minutes()) > duration {
			break
		}

		if int(currentTime.Sub(lastInvokeTime).Seconds()) > nextInterval {
			lastInvokeTime = currentTime
			nextInterval = dist.getNext(arrivalGenerator)
			log.Debug("Time to invoke! Next interval: ", nextInterval)
		}
	}
}

func (d *Driver) individualFunctionDriver(function *common.Function, announceFunctionDone *sync.WaitGroup) {
	workers := sync.WaitGroup{}

	totalTraceDuration := d.Configuration.TraceDuration

	startTime := time.Now()

	log.Debug("Spawning workers...\n")

	for i := 0; i < 1; i++ {
		workers.Add(1)
		go d.invokeFunction(startTime, totalTraceDuration)
	}

	workers.Wait()

	log.Debugf("All the invocations for function %s have been completed.\n", function.Name)
	announceFunctionDone.Done()
}

func (d *Driver) internalRun(iatOnly bool) {
	var successfulInvocations int64
	var failedInvocations int64
	allIndividualDriversCompleted := sync.WaitGroup{}
	allRecordsWritten := sync.WaitGroup{}
	allRecordsWritten.Add(1)

	log.Infof("Starting function invocation driver\n")
	for _, function := range d.Configuration.Functions {
		allIndividualDriversCompleted.Add(1)

		go d.individualFunctionDriver(function, &allIndividualDriversCompleted)
	}

	allIndividualDriversCompleted.Wait()

	log.Infof("Trace has finished executing function invocation driver\n")
	log.Infof("Number of successful invocations: \t%d\n", atomic.LoadInt64(&successfulInvocations))
	log.Infof("Number of failed invocations: \t%d\n", atomic.LoadInt64(&failedInvocations))
}

func (d *Driver) RunExperiment(iatOnly bool) {
	if d.Configuration.WithWarmup() {
		trace.DoStaticTraceProfiling(d.Configuration.Functions)
	}

	trace.ApplyResourceLimits(d.Configuration.Functions, d.Configuration.LoaderConfiguration.CPULimit)

	switch d.Configuration.LoaderConfiguration.Platform {
	case "Knative":
		driver.DeployFunctions(d.Configuration.Functions,
			d.Configuration.YAMLPath,
			d.Configuration.LoaderConfiguration.IsPartiallyPanic,
			d.Configuration.LoaderConfiguration.EndpointPort,
			d.Configuration.LoaderConfiguration.AutoscalingMetric)
	default:
		log.Fatal("Unsupported platform.")
	}

	// Generate load
	d.internalRun(iatOnly)
}