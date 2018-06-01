package traffic

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"
	"time"

	"log"

	"net/url"

	"github.com/mrcrgl/check-graceful-shutdown/pkg/cli/check-graceful-shutdown/cmd/options"
	"github.com/mrcrgl/check-graceful-shutdown/pkg/http/transport"
	"github.com/mrcrgl/check-graceful-shutdown/pkg/version"
	"github.com/pkg/errors"
)

type Simulator interface {
	Report() *SimulationReport
	Simulate(ctx context.Context, group *sync.WaitGroup)
}

func NewSimulatorForConfig(cfg options.TrafficConfig, projectName string) (*simulator, error) {
	client := &http.Client{
		Transport: &transport.UserAgent{
			Transport: http.DefaultTransport,
			UserAgent: fmt.Sprintf("%s/%s traffic-simulator", projectName, version.Version),
		},
		Timeout: cfg.RequestTimeout,
	}

	u, err := url.Parse(cfg.Target.String())
	if err != nil {
		return nil, errors.Wrapf(err, "unable to parse url %s", cfg.Target.String())
	}

	return NewSimulator(client, u, "GET", cfg.RequestConcurrency, cfg.BodyReadDelay)
}

func NewSimulator(client *http.Client, target *url.URL, method string, concurrency int, bodyReadDelay time.Duration) (*simulator, error) {

	return &simulator{
		concurrentRequests: concurrency,
		target:             target,
		method:             method,
		client:             client,
		bodyReadDelay:      bodyReadDelay,
		report:             NewSimulationReport(),
	}, nil
}

type simulator struct {
	concurrentRequests int
	bodyReadDelay      time.Duration
	target             *url.URL
	method             string
	client             *http.Client
	report             *SimulationReport
}

func (s *simulator) Test(ctx context.Context) error {
	return nil
}

func (s *simulator) Report() *SimulationReport {
	return s.report
}

func (s *simulator) Simulate(ctx context.Context, group *sync.WaitGroup) {
	group.Add(s.concurrentRequests)

	log.Printf("Start traffic simulation to %s with concurrency of %d.", s.target.String(), s.concurrentRequests)

	for n := 0; n < s.concurrentRequests; n++ {
		go func() {
			s.runSingle(ctx)
			group.Done()
		}()
	}
}

func (s *simulator) runSingle(ctx context.Context) {
loop:
	for {
		select {
		case <-ctx.Done():
			break loop
		default:
			statusCode, dur, err := s.performRequest()
			s.report.Record(statusCode, dur, err)
		}
	}
}

func (s *simulator) performRequest() (statusCode int, dur time.Duration, err error) {
	start := time.Now()
	req, err := http.NewRequest(s.method, s.target.String(), nil)
	if err != nil {
		return
	}

	req.RemoteAddr = s.target.Host
	res, err := s.client.Do(req)
	if err != nil {
		return
	}

	<-time.After(s.bodyReadDelay)
	if _, err = ioutil.ReadAll(res.Body); err != nil {
		err = errors.Wrap(err, "failed to read body")
	}

	return res.StatusCode, time.Now().Sub(start), err
}
