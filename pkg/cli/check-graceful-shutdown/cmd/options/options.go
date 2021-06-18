package options

import (
	"net/url"
	"time"
)

const ProjectName = "check-graceful-shutdown"

func NewConfigWithDefaults() *Config {
	cfg := new(Config)

	cfg.Traffic.Target.Val = url.URL{Path: "/", Host: ":8080", Scheme: "http"}
	cfg.Traffic.RequestConcurrency = 2
	cfg.Traffic.RequestTimeout = time.Second * 60
	cfg.Traffic.BodyReadDelay = time.Second * 5

	cfg.LivenessProbe.Target.Val = url.URL{Path: "/health", Host: ":8080", Scheme: "http"}
	cfg.LivenessProbe.SuccessThreshold = 1
	cfg.LivenessProbe.FailureThreshold = 3
	cfg.LivenessProbe.RequestTimeout = time.Second * 1
	cfg.LivenessProbe.InitialDelay = time.Second * 0
	cfg.LivenessProbe.Period = time.Second * 10

	cfg.ReadinessProbe.Target.Val = url.URL{Path: "/health/readiness", Host: ":8080", Scheme: "http"}
	cfg.ReadinessProbe.SuccessThreshold = 1
	cfg.ReadinessProbe.FailureThreshold = 3
	cfg.ReadinessProbe.RequestTimeout = time.Second * 1
	cfg.ReadinessProbe.InitialDelay = time.Second * 0
	cfg.ReadinessProbe.Period = time.Second * 2

	return cfg
}

type Config struct {
	ProjectName    string
	LivenessProbe  ProbeConfig
	ReadinessProbe ProbeConfig
	Traffic        TrafficConfig
	Process        ProcessConfig
}

type ProcessConfig struct {
	PID       int
	Command   string
	Arguments []string
}

type TrafficConfig struct {
	Target             URI
	RequestConcurrency int
	RequestTimeout     time.Duration
	BodyReadDelay      time.Duration
}

type ProbeConfig struct {
	RequestTimeout   time.Duration
	Target           URI
	InitialDelay     time.Duration
	Period           time.Duration
	SuccessThreshold int
	FailureThreshold int
}

type URI struct {
	Val url.URL
}

func (u *URI) String() string {
	return u.Val.String()
}

func (u *URI) Set(value string) error {
	uri, err := url.Parse(value)
	if err != nil {
		return err
	}
	u.Val = *uri

	return nil
}

func (u *URI) Type() string {
	return "url"
}

func CutProcessConfigFromArgs(args ...string) ([]string, ProcessConfig) {
	pc := ProcessConfig{}

	var retainedArgs []string
	var dividerPos int

loop:
	for n, argument := range args {
		switch {
		case argument == "--":
			dividerPos = n
			retainedArgs = args[0:n]
		case dividerPos != 0 && n-dividerPos == 1:
			pc.Command = argument
		case dividerPos != 0 && n-dividerPos > 1:
			pc.Arguments = args[n:len(args)]
			break loop
		}
	}

	return retainedArgs, pc
}
