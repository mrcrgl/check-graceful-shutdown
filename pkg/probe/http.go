package probe

import (
	"context"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"sync"

	"fmt"

	"net/url"

	"github.com/fid-dev/check-graceful-shutdown/pkg/cli/check-graceful-shutdown/cmd/options"
	"github.com/fid-dev/check-graceful-shutdown/pkg/http/transport"
	"github.com/fid-dev/check-graceful-shutdown/pkg/version"
	"github.com/pkg/errors"
)

func NewHTTPForConfig(cfg options.ProbeConfig, projectName string) (*httpProbe, error) {

	client := &http.Client{
		Timeout: cfg.RequestTimeout,
		Transport: &transport.UserAgent{
			Transport: http.DefaultTransport,
			UserAgent: fmt.Sprintf("%s/%s http-probe", projectName, version.Version),
		},
	}

	u, err := url.Parse(cfg.Target.String())
	if err != nil {
		return nil, errors.Wrapf(err, "unable to parse url %s", cfg.Target.String())
	}

	return NewHTTP(
		client,
		u,
		cfg.InitialDelay,
		cfg.Period,
		cfg.SuccessThreshold,
		cfg.FailureThreshold,
	)
}

func NewHTTP(client *http.Client, target *url.URL, initialDelay, period time.Duration, successThreshold, failureThreshold int) (*httpProbe, error) {

	return &httpProbe{
		initialDelay:     initialDelay,
		period:           period,
		successThreshold: successThreshold,
		failureThreshold: failureThreshold,
		target:           target,
		client:           client,
		bucket:           make([]Status, 10),
		subscribers:      make([]chan Status, 0),
	}, nil
}

var _ Interface = &httpProbe{}

type httpProbe struct {
	client           *http.Client
	initialDelay     time.Duration
	period           time.Duration
	Timeout          time.Duration
	successThreshold int
	failureThreshold int
	target           *url.URL
	status           Status
	bucket           []Status
	bucketMu         sync.Mutex
	subscribers      []chan Status
}

func (h *httpProbe) Check() error {
	if h.Timeout > h.period {
		return errors.Errorf("probe timeout of %s must be lower than period %s", h.Timeout.String(), h.period.String())
	}

	return nil
}

func (h *httpProbe) Run(ctx context.Context) {
	_ = <-time.After(h.initialDelay)

loop:
	for {
		select {
		case <-time.After(h.period):
			go h.check()
		case <-ctx.Done():
			log.Printf("probe closed by context with: %s", ctx.Err())
			break loop
		}
	}
}

func (h *httpProbe) Notify(sCh chan Status) {
	h.subscribers = append(h.subscribers, sCh)
}

func (h *httpProbe) check() {
	req, err := http.NewRequest("GET", h.target.Path, nil)
	if err != nil {
		h.pushStatus(Failure, errors.Wrap(err, "failed to build request"))
		return
	}

	req.RemoteAddr = h.target.Host

	res, err := h.client.Do(req)
	if err != nil {
		h.pushStatus(Failure, err)
		return
	}

	if _, err := ioutil.ReadAll(res.Body); err != nil {
		h.pushStatus(Failure, errors.Wrap(err, "failed to read response"))
		return
	}

	if res.StatusCode < http.StatusOK || res.StatusCode >= http.StatusBadRequest {
		h.pushStatus(Failure, errors.Errorf("bad response code: %d", res.StatusCode))
		return
	}

	h.pushStatus(Success, nil)
}

func (h *httpProbe) pushStatus(status Status, err error) {
	h.bucketMu.Lock()
	defer h.bucketMu.Unlock()

	bucket := make([]Status, 0)
	bucket = append(bucket, status)
	bucket = append(bucket, h.bucket[0:9]...)
	h.bucket = bucket

	log.Printf("Bucket: %+v", h.bucket)
	h.evalStatus()
}

func (h *httpProbe) evalStatus() {
	if len(h.bucket[0]) == 0 {
		return
	}

	nextStatus := h.bucket[0]

	if nextStatus == h.status {
		return
	}

	var threshold int
	switch nextStatus {
	case Success:
		threshold = h.successThreshold
	case Failure:
		threshold = h.failureThreshold
	default:
		log.Printf("unknown next status: %s", nextStatus)
		return
	}

	var consecutive int
	for _, status := range h.bucket {
		if status == nextStatus {
			consecutive++
		} else {
			break
		}
	}

	if consecutive >= threshold {
		h.setStatus(nextStatus)
	}
}

func (h *httpProbe) setStatus(status Status) {
	if h.status == status {
		return
	}

	h.status = status
	h.notifySubscribers(status)
}

func (h *httpProbe) notifySubscribers(status Status) {
	for _, sCh := range h.subscribers {
		go func(ch chan Status) {
			ch <- status
		}(sCh)
	}
}
