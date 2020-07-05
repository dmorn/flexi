package fargate

import (
	"fmt"

	"github.com/jecoz/flexi"
)

type Fargate struct {
}

func New() *Fargate {
	return new(Fargate)
}

func (f *Fargate) Spawn(t flexi.Task) (*flexi.RemoteProcess, error) {
	return nil, fmt.Errorf("not implemented yet")
}
