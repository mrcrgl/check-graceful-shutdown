package traffic

import "time"

type SimulationReport struct {
}

func (sr *SimulationReport) Record(statusCode int, elapsedTime time.Duration, err error) {

}

func (sr *SimulationReport) String() string {
	return "<report>"
}
