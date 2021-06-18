package grace

import (
	"context"
	"log"
	"sync"

	"time"

	"github.com/mrcrgl/check-graceful-shutdown/pkg/cli/check-graceful-shutdown/cmd/options"
	"github.com/mrcrgl/check-graceful-shutdown/pkg/probe"
	"github.com/mrcrgl/check-graceful-shutdown/pkg/process"
	"github.com/mrcrgl/check-graceful-shutdown/pkg/traffic"
	"github.com/pkg/errors"
)

func NewConductor(cfg *options.Config) (*Conductor, error) {
	liveness, err := probe.NewHTTPForConfig(cfg.LivenessProbe, probe.Success)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create liveness probe")
	}

	readiness, err := probe.NewHTTPForConfig(cfg.ReadinessProbe, probe.Failure)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create readiness probe")
	}

	simulator, err := traffic.NewSimulatorForConfig(cfg.Traffic)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create traffic simulator")
	}

	handler := process.NewHandler(cfg.Process.Command, cfg.Process.Arguments...)

	return &Conductor{
		processHandler: handler,
		livenessProbe:  liveness,
		readinessProbe: readiness,
		traffic:        simulator,
	}, nil
}

type LifecycleStatus int

type Conductor struct {
	processHandler process.Handler
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
	processCh := make(chan process.Status)

	c.livenessProbe.Notify(livenessCh)
	c.readinessProbe.Notify(readinessCh)
	c.processHandler.Notify(processCh)

	trafficCtx, trafficCancel := context.WithCancel(ctx)

	wg := new(sync.WaitGroup)

	go func() {
		ctxProbes, cancelProbes := context.WithCancel(ctx)

	loop:
		for {
			select {
			case procStatus := <-processCh:
				log.Printf("process status changed to %s\n", procStatus)
				switch procStatus {
				case process.Running:
					go c.livenessProbe.Run(ctxProbes)
					go c.readinessProbe.Run(ctxProbes)
				case process.Exited:
					trafficCancel()
					cancelProbes()
					break loop
				}
			}
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		if err := c.processHandler.Start(ctx); err != nil {
			log.Panicf("failed to handle process: %s", err)
		}
	}()

	go func() {
	loop:
		for {
			select {
			case status := <-livenessCh:
				log.Printf("liveness status changed to %s\n", status)
				if status == probe.Failure {
					break loop
				}
			case status := <-readinessCh:
				log.Printf("readiness status changed to %s\n", status)
				if status == probe.Success {
					go c.traffic.Simulate(trafficCtx, wg)
					<-time.After(time.Second * 10)
					go c.initiateShutdown()
				} else {
					trafficCancel()
				}
			}
		}
	}()

	wg.Wait()
}

func (c *Conductor) initiateShutdown() {
	processCh := make(chan process.Status)

	c.processHandler.Notify(processCh)
	c.processHandler.Signal(process.SignalTerminate)

	select {
	case s := <-processCh:
		if s == process.Exited {
			break
		}
	case <-time.After(time.Second * 30):
		c.processHandler.Signal(process.SignalKill)
	}
}

func (c *Conductor) Report() *traffic.SimulationReport {
	return c.traffic.Report()
}
