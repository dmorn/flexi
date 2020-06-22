package main

import (
	"fmt"

	"github.com/jecoz/flexi"
)

const (
	envRegistrationAddress = "FLEXI_REGISTRATION_ADDRESS"
	envProcessName         = "FLEXI_PROCESS_NAME"
	envTaskId              = "FLEXI_TASK_ID"
)

// Retrieves registration address, process name and task identifier
// from environment variables.
func NewProcessFromEnv() (*flexi.Process, error) {
	reg, ok := os.LookupEnv(envRegistrationAddress)
	if !ok {
		return nil, fmt.Errorf("%v not set", envRegistrationAddress)
	}
	u := url.Parse(reg)
	if u.Scheme != "amqp" {
		return nil, fmt.Errorf("unsupported registration scheme %v", u.Scheme)
	}

	name, ok := os.LookupEnv(envProcessName)
	if !ok {
		return nil, fmt.Errorf("%v not set", envProcessName)
	}
	id, ok := os.LookupEnv(envTaskId)
	if !ok {
		return nil, fmt.Errorf("%v not set", envTaskId)
	}

	t, err := rabbit.NewTransportPassive(u)
	if err != nil {
		return nil, err
	}
	return &flexi.Process{
		Transporter: t,
		Name: name,
		TaskId: id,
	}, nil
}

type Params struct {
	Name string `json:"name"`
}

func main() {
	p, err := NewProcessFromEnv()
	if err != nil {
		panic(err)
	}
	defer p.Close()

	var params Params
	if err := p.Register(&params); err != nil {
		panic(err)
	}
	defer p.OK(struct{}{})

	fmt.Printf("hello from ** %v **\n", params.Name)
}
