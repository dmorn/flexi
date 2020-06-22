package flexi

import (
	"context"
	"fmt"
)

type Closer interface {
	Close() error
}

type Receiver interface {
	Recv(context.Context) (*Msg, error)
}

type Publisher interface {
	Publish(*Msg) error
}

type Transporter interface {
	Publisher
	Receiver
	Closer
}

// In docker terms, a Process would be a container, hence the runtime
// representation of the image, once it is executed.
// Every exported field should be set before calling any function.
type Process struct {
	Name   string
	TaskId string
	Transporter
}

func (p *Process) Publish(t string, v interface{}) error {
	msg, err := MarshalMsg(t, v)
	if err != nil {
		return err
	}
	return p.Transporter.Publish(msg)
}

type RegistrationRequest struct {
	TaskId      string `json:"task_id"`
	ProcessName string `json:"process_name"`
}

// Register exchanges a registration message with the other side,
// retrieving the parameters required for starting the task.
func (p *Process) Register(ctx context.Context, params interface{}) error {
	if p.TaskId == "" {
		return fmt.Errorf("task identifier not set")
	}

	// TODO: add connection deadlines before getting stuck reading / writing
	// messages.

	if err := p.Publish("registration", &RegistrationRequest{
		TaskId: p.TaskId,
	}); err != nil {
		return fmt.Errorf("publish registration message: %w", err)
	}

	// Now wait for the other side to publish the registration message.
	// No other message will be accepted at this stage.
	msg, err := p.Recv(ctx)
	if err != nil {
		return fmt.Errorf("receiving registration message: %w", err)
	}
	if err := msg.Unmarshal(params); err != nil {
		return fmt.Errorf("receiving registration message: %w", err)
	}

	return nil
}

// OK publishes an ok message. Usually denotes that the process
// successfully completed its task.
func (p *Process) OK(payload interface{}) error {
	return p.Publish("ok", payload)
}

// Err publishes an error message. Usually denotes a processing failure that
// cannot be recovered.
func (p *Process) Err(err error) error {
	return p.Publish("error", &struct {
		Error string `json:"error"`
	}{
		Error: err.Error(),
	})
}

type RemoteProcess struct {
	Addr  string `json:"addr"`
	Queue string `json:"queue"`
}
