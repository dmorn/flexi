package fargate

import (
	"context"
	"testing"
	"time"

	"github.com/jecoz/flexi"
)

func TestSpawn(t *testing.T) {
	fargate := &Fargate{
		SecurityGroups: []string{
			"sg-01a4bacf92d52fc75",
		},
		Subnets: []string{
			"subnet-0f3a619bdcac8cd3c",
			"subnet-052c680353e0dd406",
		},
		Cluster: "tooling",
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*5)
	defer cancel()
	remote, err := fargate.Spawn(ctx, flexi.Task{
		ID: "test",
		Image: &flexi.Image{
			Type:    "ecs-task",
			Name:    "echo64",
			Service: "9pfs",
		},
		Caps: &flexi.Caps{},
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("remote process address: %v", remote.Addr())
	t.Fatal("inspection time")
}
