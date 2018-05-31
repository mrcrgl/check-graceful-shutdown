package grace

import (
	"context"

	"sync"

	"github.com/fid-dev/check-graceful-shutdown/pkg/cli/check-graceful-shutdown/cmd/options"
	"github.com/fid-dev/check-graceful-shutdown/pkg/probe"
	"github.com/fid-dev/check-graceful-shutdown/pkg/traffic"
	"github.com/pkg/errors"
)

func NewConductor(cfg *options.Config) (*Conductor, error) {
	liveness, err := probe.NewHTTPForConfig(cfg.LivenessProbe, cfg.ProjectName)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create liveness probe")
	}

	readiness, err := probe.NewHTTPForConfig(cfg.ReadinessProbe, cfg.ProjectName)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create readiness probe")
	}

	simulator, err := traffic.NewSimulatorForConfig(cfg.Traffic)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create traffic simulator")
	}

	return &Conductor{
		livenessProbe:  liveness,
		readinessProbe: readiness,
		traffic:        simulator,
	}, nil
}

type LifecycleStatus int

type Conductor struct {
	livenessProbe  probe.Interface
	readinessProbe probe.Interface
	traffic        traffic.Simulator
}

/*
start both probes

start to flow traffic when readiness is green
stop to flow traffic when readiness is red
capture errors from simulator and prepare report
*/

// Run starts the test. The function blocks until the test is done. To abort, cancel the context.
func (c *Conductor) Run(ctx context.Context) {
	livenessCh := make(chan probe.Status)
	readinessCh := make(chan probe.Status)

	c.livenessProbe.Notify(livenessCh)
	c.readinessProbe.Notify(readinessCh)

	go c.livenessProbe.Run(ctx)
	go c.readinessProbe.Run(ctx)

	trafficCtx, trafficCancel := context.WithCancel(ctx)

	wg := new(sync.WaitGroup)

loop:
	for {
		select {
		case status := <-livenessCh:
			if status == probe.Failure {
				break loop
			}
		case status := <-readinessCh:
			if status == probe.Success {
				go c.traffic.Simulate(trafficCtx, wg)
			} else {
				trafficCancel()
			}
		}
	}

	wg.Wait()
}

func (c *Conductor) Report() *traffic.SimulationReport {
	return c.traffic.Report()
}
