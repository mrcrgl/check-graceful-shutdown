package cmd

import (
	"errors"
	"fmt"

	"os"

	"context"
	"os/signal"
	"syscall"

	"log"

	"github.com/mrcrgl/check-graceful-shutdown/pkg/cli/check-graceful-shutdown/cmd/options"
	"github.com/mrcrgl/check-graceful-shutdown/pkg/grace"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func NewRootCommand() *cobra.Command {
	cfg := options.NewConfigWithDefaults()

	var root = &cobra.Command{
		Use:   cfg.ProjectName,
		Short: "tool to check if a service supports graceful shutdown",
		Long:  ``,
		Run: func(cmd *cobra.Command, args []string) {
			if cfg.Process.Command == "" {
				fail(errors.New("invalid command to execute"))
			}

			c, err := grace.NewConductor(cfg)
			if err != nil {
				fail(err)
			}

			ctx, cancel := context.WithCancel(context.Background())

			go func() {
				sigCh := make(chan os.Signal)

				signal.Notify(sigCh, syscall.SIGTERM, os.Interrupt)

				var isTerminated bool

				for {
					select {
					case s := <-sigCh:
						if isTerminated {
							log.Println("cancelled")
							os.Exit(128 + 143)
						} else {
							log.Printf("User signal %s\n", s)
							isTerminated = true
							cancel()
						}
					}
				}
			}()

			c.Run(ctx)

			log.Printf("Report:\n%s", c.Report().String())

			log.Println("done.")
		},
	}

	//root.Flags().IntVarP(&cfg.Process.PID, "pid", "p", 0, "pid of the process")
	//root.Flags().StringVar(&cfg.Process.Command, "exec", cfg.Process.Command, "command to execute")
	root.Flags().Var(&cfg.Traffic.Target, "traffic-target", "http endpoint to simulate traffic to")
	root.Flags().IntVar(&cfg.Traffic.RequestConcurrency, "traffic-request-concurrency", cfg.Traffic.RequestConcurrency, "number of concurrent requests to perform")
	root.Flags().DurationVar(&cfg.Traffic.RequestTimeout, "traffic-request-timeout", cfg.Traffic.RequestTimeout, "http request timeout")

	addProbeFlags(root.Flags(), "liveness", &cfg.LivenessProbe)
	addProbeFlags(root.Flags(), "readiness", &cfg.ReadinessProbe)

	var args []string
	args, cfg.Process = options.CutProcessConfigFromArgs(os.Args...)

	root.ParseFlags(args)

	return root
}

func fail(err error) {
	log.Printf("An error occurred.\nError: %s\n", err)
	os.Exit(1)
}

func addProbeFlags(fs *pflag.FlagSet, kind string, pc *options.ProbeConfig) {
	fs.DurationVar(
		&pc.Period,
		fmt.Sprintf("%s-probe-period", kind),
		pc.Period,
		fmt.Sprintf("period of %s checks", kind),
	)

	fs.DurationVar(
		&pc.InitialDelay,
		fmt.Sprintf("%s-probe-initial-delay", kind),
		pc.InitialDelay,
		fmt.Sprintf("initial delay before starting with %s checks", kind),
	)

	fs.DurationVar(
		&pc.InitialDelay,
		fmt.Sprintf("%s-probe-request-timeout", kind),
		pc.InitialDelay,
		fmt.Sprintf("http timeout for %s checks", kind),
	)

	fs.IntVar(
		&pc.FailureThreshold,
		fmt.Sprintf("%s-probe-failure-threshold", kind),
		pc.FailureThreshold,
		fmt.Sprintf("consequent number of %s checks required to fail", kind),
	)

	fs.IntVar(
		&pc.SuccessThreshold,
		fmt.Sprintf("%s-probe-success-threshold", kind),
		pc.SuccessThreshold,
		fmt.Sprintf("consequent number of %s checks required to succeed", kind),
	)

	fs.Var(
		&pc.Target,
		fmt.Sprintf("%s-probe-target", kind),
		fmt.Sprintf("http endpoint to perform %s checks", kind),
	)
}
