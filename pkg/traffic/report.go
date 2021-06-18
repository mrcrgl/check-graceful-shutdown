package traffic

import (
	"bytes"
	"fmt"
	"sync"
	"time"
)

func NewSimulationReport() *SimulationReport {
	return &SimulationReport{
		httpCodes: make(httpCodesVec),
		errors:    make([]error, 0),
	}
}

type SimulationReport struct {
	mu        sync.RWMutex
	httpCodes httpCodesVec
	errors    []error
}

func (sr *SimulationReport) Record(statusCode int, elapsedTime time.Duration, err error) {
	sr.mu.Lock()
	defer sr.mu.Unlock()

	if err != nil {
		sr.errors = append(sr.errors, err)
	}

	if statusCode > 0 {
		sr.httpCodes[fmt.Sprintf("%d", statusCode)]++
	}
}

func (sr *SimulationReport) String() string {
	sr.mu.RLock()
	defer sr.mu.RUnlock()

	buf := bytes.NewBuffer([]byte(""))

	fmt.Fprint(buf, "http response code count:\n")
	for code, num := range sr.httpCodes {
		fmt.Fprintf(buf, "\t%s: %d\n", code, num)
	}

	fmt.Fprint(buf, "\n")

	fmt.Fprintf(buf, "num errors: %d", len(sr.errors))

	if len(sr.errors) == 0 {
		fmt.Fprint(buf, "GRACEFUL SHUTDOWN SUCCEED\n")
	} else {
		fmt.Fprintf(buf, "GRACEFUL SHUTDOWN FAILED WITH %d ERRORS!\n", len(sr.errors))
	}

	return buf.String()
}

type httpCodesVec map[string]int
