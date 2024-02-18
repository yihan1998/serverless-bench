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

func (d *Driver) invokeFunction() {
	// defer metadata.AnnounceDoneWG.Done()

	// var success bool

	// var record *mc.ExecutionRecord
	// switch d.Configuration.LoaderConfiguration.Platform {
	// case "Knative":
	// 	success, record = InvokeGRPC(
	// 		metadata.Function,
	// 		metadata.RuntimeSpecifications,
	// 		d.Configuration.LoaderConfiguration,
	// 	)
	// default:
	// 	log.Fatal("Unsupported platform.")
	// }

	// record.Phase = int(metadata.Phase)
	// record.InvocationID = composeInvocationID(d.Configuration.TraceGranularity, metadata.MinuteIndex, metadata.InvocationIndex)

	// metadata.RecordOutputChannel <- record

	// if success {
	// 	atomic.AddInt64(metadata.SuccessCount, 1)
	// } else {
	// 	atomic.AddInt64(metadata.FailedCount, 1)
	// 	atomic.AddInt64(&metadata.FailedCountByMinute[metadata.MinuteIndex], 1)
	// }
	log.Debug("Invoking function...")
}

func (d *Driver) individualFunctionDriver(function *common.Function, announceFunctionDone *sync.WaitGroup) {
	waitForInvocations := sync.WaitGroup{}

	totalTraceDuration := d.Configuration.TraceDuration

	startTime := time.Now()

	for {
		currentTime := time.Now()

		if int(currentTime.Sub(startTime).Minutes()) > totalTraceDuration {
			break
		}

		waitForInvocations.Add(1)

		// go d.invokeFunction(&InvocationMetadata{
		// 	Function:              function,
		// 	RuntimeSpecifications: &runtimeSpecification[minuteIndex][invocationIndex],
		// 	Phase:                 currentPhase,
		// 	MinuteIndex:           minuteIndex,
		// 	InvocationIndex:       invocationIndex,
		// 	SuccessCount:          &successfulInvocations,
		// 	FailedCount:           &failedInvocations,
		// 	FailedCountByMinute:   failedInvocationByMinute,
		// 	RecordOutputChannel:   recordOutputChannel,
		// 	AnnounceDoneWG:        &waitForInvocations,
		// 	AnnounceDoneExe:       addInvocationsToGroup,
		// 	ReadOpenWhiskMetadata: readOpenWhiskMetadata,
		// })

		go d.invokeFunction()

		waitForInvocations.Wait()
	}

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
