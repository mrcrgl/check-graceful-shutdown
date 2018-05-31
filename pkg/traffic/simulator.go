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

	"github.com/fid-dev/check-graceful-shutdown/pkg/cli/check-graceful-shutdown/cmd"
	"github.com/fid-dev/check-graceful-shutdown/pkg/http/transport"
	"github.com/fid-dev/check-graceful-shutdown/pkg/version"
	"github.com/pkg/errors"
)

type Simulator interface {
	Report() *SimulationReport
	Simulate(ctx context.Context, group *sync.WaitGroup)
}

func NewSimulatorForConfig(cfg cmd.TrafficConfig) (*simulator, error) {
	return NewSimulator(cfg.Target.String(), "GET", cfg.RequestConcurrency, cfg.RequestTimeout)
}

func NewSimulator(target string, method string, concurrency int, timeout time.Duration) (*simulator, error) {
	client := &http.Client{
		Transport: &transport.UserAgent{
			Transport: http.DefaultTransport,
			UserAgent: fmt.Sprintf("%s/%s traffic-simulator", cmd.Name, version.Version),
		},
		Timeout: timeout,
	}

	u, err := url.Parse(target)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to parse url %s", target)
	}

	return &simulator{
		concurrentRequests: concurrency,
		target:             u,
		method:             method,
		client:             client,
	}, nil
}

type simulator struct {
	concurrentRequests int
	target             *url.URL
	method             string
	client             *http.Client
}

func (s *simulator) Test(ctx context.Context) error {
	return nil
}

func (s *simulator) Report() *SimulationReport {
	return &SimulationReport{}
}

func (s *simulator) Simulate(ctx context.Context, group *sync.WaitGroup) {
	group.Add(s.concurrentRequests)

	log.Printf("Start traffic simulation to %s with concurrency of %d.", s.target.String(), s.concurrentRequests)

	for n := 0; n <= s.concurrentRequests; n++ {
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
			if err := s.performRequest(); err != nil {
				// TODO handle error
				log.Printf("request failed with: %s", err)
			}
		}
	}
}

func (s *simulator) performRequest() error {
	req, err := http.NewRequest(s.method, s.target.Path, nil)
	if err != nil {
		return err
	}

	req.RemoteAddr = s.target.Host
	res, err := s.client.Do(req)
	if err != nil {
		return err
	}

	if _, err := ioutil.ReadAll(res.Body); err != nil {
		return errors.Wrap(err, "failed to read body")
	}

	if res.StatusCode < http.StatusOK || res.StatusCode >= http.StatusBadRequest {
		return errors.Errorf("unexpected status code: %d", res.StatusCode)
	}

	return nil
}
