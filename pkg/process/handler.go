package process

import (
	"context"
	"log"
	"os/exec"

	"os"

	"time"

	"syscall"

	"github.com/pkg/errors"
)

type Handler interface {
	Start(ctx context.Context) error
	Signal(signal Signal)
	Notify(sCh chan Status)
}

type Status string

const (
	Running Status = "running"
	Exited         = "exited"
)

type Signal string

const (
	SignalTerminate Signal = "terminate"
	SignalKill             = "kill"
)

func NewHandler(cmd string, args ...string) *handler {
	return &handler{
		subscribers: make([]chan Status, 0),
		cCh:         make(chan Signal),
		cmd:         cmd,
		args:        args,
		status:      Exited,
	}
}

var _ Handler = &handler{}

type handler struct {
	status      Status
	subscribers []chan Status
	cCh         chan Signal
	cmd         string
	args        []string
}

// start process, (start listener and traffic flow), receive SIGINT, wait for process to exit, kill process after 30s
func (h *handler) Start(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, h.cmd, h.args...)

	cmd.Env = os.Environ()
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout

	go func() {
		for {
			select {
			case <-time.After(time.Millisecond * 100):
				if cmd.ProcessState != nil && cmd.ProcessState.Exited() {
					h.setStatus(Exited)
				} else {
					h.setStatus(Running)
				}
			case sig := <-h.cCh:
				switch sig {
				case SignalTerminate:
					if cmd.ProcessState == nil || !cmd.ProcessState.Exited() {
						if err := cmd.Process.Signal(syscall.SIGTERM); err != nil {
							log.Printf("failed to send signal %s to process pid=%d: %s", SignalTerminate, cmd.Process.Pid, err)
						}
					}
				case SignalKill:
					if cmd.ProcessState == nil || !cmd.ProcessState.Exited() {
						if err := cmd.Process.Kill(); err != nil {
							log.Printf("failed to kill process pid=%d: %s", cmd.Process.Pid, err)
						}
					}
				}
			}
		}
	}()

	if err := cmd.Run(); err != nil {
		return errors.Wrap(err, "command execution failed")
	}

	return nil
}

func (h *handler) Signal(signal Signal) {
	h.cCh <- signal
}

func (h *handler) Notify(sCh chan Status) {
	h.subscribers = append(h.subscribers, sCh)
}

func (h *handler) setStatus(status Status) {
	if h.status == status {
		return
	}

	h.status = status
	h.notifySubscribers(status)
}

func (h *handler) notifySubscribers(status Status) {
	for _, sCh := range h.subscribers {
		go func(ch chan Status) {
			ch <- status
		}(sCh)
	}
}
