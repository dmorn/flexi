package fargate

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/jecoz/flexi"
)


type Fargate struct {
	svc *ecs.ECS
}

// New returns an authenticated client that can use fargate
// services. To authenticate, configure and such we rely on
// environment variables.
func New() *Fargate {
	sess := session.Must(session.NewSession())
	return &Fargate{
		svc: ecs.New(sess),
	}
}

func (f *Fargate) Spawn(ctx context.Context, t *flexi.Task, r flexi.Registration) (*flexi.Process, error) {
	return nil, fmt.Errorf("not implemented yet")
}
