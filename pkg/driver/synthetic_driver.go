package driver

import (
	"sync"

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

func (d *Driver) individualFunctionDriver(function *common.Function, announceFunctionDone *sync.WaitGroup,
	addInvocationsToGroup *sync.WaitGroup, readOpenWhiskMetadata *sync.Mutex, totalSuccessful *int64,
	totalFailed *int64, totalIssued *int64, recordOutputChannel chan interface{}) {

}

func (d *Driver) internalRun(iatOnly bool) {
	log.Infof("Starting function invocation driver\n")
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
