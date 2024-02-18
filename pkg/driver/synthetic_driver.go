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

	dist "github.com/yihan1998/serverless-bench/pkg/distribution"
)

type DriverConfiguration struct {
	LoaderConfiguration *config.LoaderConfiguration
	IATDistribution     common.IatDistribution
	ShiftIAT            bool // shift the invocations inside minute
	TraceGranularity    common.TraceGranularity
	TraceDuration       int // in minutes

	Rate      float64
	NumWorker int

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

func (d *Driver) invokeFunction() {

}

func (d *Driver) workerRoutine(rate float64, duration int) {
	var arrivalGenerator = dist.NewExponentialGenerator(rate)
	var nextInterval = arrivalGenerator.GetNext()
	invokedFunctions := sync.WaitGroup{}

	numberOfInvocations := 0
	perSecInvocations := 0

	log.Debug("Duration: ", duration)

	startTime := time.Now()
	lastInvokeTime := time.Now()
	lastLogTime := time.Now()

	for {
		totalElapsed := time.Since(startTime)
		invokeElapsed := time.Since(lastInvokeTime)
		lastLogElapsed := time.Since(lastLogTime)

		if int(totalElapsed.Minutes()) > duration {
			break
		}

		if int(lastLogElapsed.Seconds()) > 1 {
			numberOfInvocations += perSecInvocations
			log.Debug("Rate: ", perSecInvocations/int(lastLogElapsed.Milliseconds()), "(KRPS)")
			lastLogTime = time.Now()
			break
		}

		if invokeElapsed.Milliseconds() > nextInterval {
			invokedFunctions.Add(1)
			go d.invokeFunction()

			perSecInvocations += 1
			lastInvokeTime = time.Now()
			nextInterval = arrivalGenerator.GetNext()
		}
	}

	invokedFunctions.Wait()
	totalTime := time.Since(startTime)

	log.Infof("Experiment took %s, request rate: %v (KRPS)", totalTime, numberOfInvocations/int(totalTime.Milliseconds()))
}

func (d *Driver) individualFunctionDriver(function *common.Function, announceFunctionDone *sync.WaitGroup) {
	workerGroup := sync.WaitGroup{}

	totalTraceDuration := d.Configuration.TraceDuration

	per_worker_rate := d.Configuration.Rate / float64(d.Configuration.NumWorker)

	for i := 0; i < d.Configuration.NumWorker; i++ {
		workerGroup.Add(1)
		go d.workerRoutine(per_worker_rate, totalTraceDuration)
	}

	workerGroup.Wait()

	log.Debugf("All the invocations for function %s have been completed.\n", function.Name)
	announceFunctionDone.Done()
}

func (d *Driver) internalRun(iatOnly bool) {
	var successfulInvocations int64
	var failedInvocations int64
	allIndividualDriversCompleted := sync.WaitGroup{}
	allRecordsWritten := sync.WaitGroup{}
	allRecordsWritten.Add(1)

	log.Infof("Starting function invocation driver(%v worker @%v KRPS)\n", d.Configuration.NumWorker, d.Configuration.Rate)
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
