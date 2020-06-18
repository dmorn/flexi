package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"
)

func errorf(format string, args ...interface{}) {
	fmt.Printf("error * "+format+"\n", args...)
}

func logf(format string, args ...interface{}) {
	fmt.Printf(format+"\n", args...)
}

// Image describes what will be executed.
type Image struct {
	// Type is usually docker.
	Type string `json:"type"`
	// Name is usually an address to a docker image stored
	// in some registry.
	Name string `json:"name"`
}

// Based on the required capabilities, we'll choose where the
// container should be executed.
type Caps struct {
	CPU int `json:"cpu"`
	Ram int `json:"ram"`
	GPU int `json:"gpu"`
}

// Task defines **what** should be executed, on **which** hardware.
type Task struct {
	ID    string `json:"id"`
	Image *Image `json:"image"`
	Caps  *Caps  `json:"capabilities"`
}

// Conn is used as a bridge to a process.
type Conn struct {
	Proto string
	Addr  string
}

// In docker terms, a Process would be a container, hence the runtime
// representation of the image, once it is executed.
type Process struct {
	ID   string
	Conn *Conn
}

// Registration is performed right after the process is started, but before
// it has ensured the caller that it is ready to start the task. When used,
// a nil error indicated that the process is bound to the conn and its ready
// to be used. The connection should be passed over to the requesting entity.
type Registration func(context.Context, *Process, *Conn) error

// Given a task definition, Spawner is capable of instantiating the
// image into a process, respecting the required hardware capabilities.
type Spawner interface {
	Spawn(context.Context, *Task, Registration) (*Process, error)
}

type FakeSpawner struct {
}

func (r *FakeSpawner) Spawn(ctx context.Context, task *Task, register Registration) (*Process, error) {
	p := &Process{ID: "faker"}
	conn := &Conn{Proto: "amqp", Addr: "localhost:6667"}
	if err := register(ctx, p, conn); err != nil {
		return nil, err
	}
	p.Conn = conn
	return p, nil
}

func nopRegister(ctx context.Context, p *Process, c *Conn) error {
	return nil
}

func main() {
	var task Task
	if err := json.NewDecoder(os.Stdin).Decode(&task); err != nil {
		errorf("decode task error: %v", err)
		return
	}

	logf("decoded task [%v]", task.ID)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	process, err := new(FakeSpawner).Spawn(ctx, &task, nopRegister)
	if err != nil {
		errorf(err.Error())
		return
	}
	logf("started process [%v]", process.ID)
}
