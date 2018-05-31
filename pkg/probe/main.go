package probe

import "context"

type Interface interface {
	Run(ctx context.Context)
	Check() error
	Notify(sCh chan Status)
}
