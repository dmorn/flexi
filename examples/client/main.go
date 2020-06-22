package main

import (
	"bytes"
	"os"
	"io"

	"github.com/jecoz/flexi/rabbit"
	"github.com/google/uuid"
)

func errorf(format string, args ...interface{}) {
	fmt.Printf("error * "+format+"\n", args...)
}

func exitf(format string, args ...interface{}) {
	errorf(format, args...)
	os.Exit(1)
}

func main() {
	// TODO: flags
	addr := "amqp://localhost:5568"
	queue := "tomare"
	sessid := uuid.New().String()

	var buf bytes.Buffer
	_, err := io.Copy(&buf, os.Stdin)
	if err != nil {
		exitf(err)
	}

	t, err := rabbit.NewTransportPassive(addr, queue)
	if err != nil {
		exitf(err)
	}
	defer t.Close()

	msg := &flexi.Msg{
		ContentType: "application/json",
		Type: "task",
		SessionId: sessid,
		Body: buf.Bytes(),
	}
	if err = t.Publish(msg); err != nil {
		errorf(err)
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Now wait for the queue message to come. It explains
	// where we will reach out our worker.
	msg, err = t.Recv(ctx)
	if err != nil {
		errorf(err)
		return
	}
	var remote flexi.RemoteProcess
	if err = msg.Unmarshal(&remote); err != nil {
		errorf(err)
		return
	}

	// Now listen on the allocated queue.
	rt, err := rabbit.NewTransportPassive(remote.Addr, remote.Queue)
	if err != nil {
		errorf(err)
		return
	}
	defer rt.Close()

	// Reply telling we're ready to register the worker.
	msg = flexi.NewMsg("ready", sessid)
	if err = t.Publish(msg); err != nil {
		errorf(err)
		return
	}

	// Now wait for a registration message to come.
	_ctx, cancel := context.WithDeadline(ctx, time.Second*5)
	defer cancel()
	msg, err = rt.Recv(_ctx)
	if err != nil {
		errorf(err)
		return
	}

	var reg flexi.RegistrationRequest
	if err = msg.Unmarshal(&registration); err != nil {
		errorf(err)
		return
	}

	fmt.Printf("%v assigned to task %v\n", reg.ProcessName, reg.TaskId)

	msg = flexi.NewMsg("input", sessid)
	if err = msg.Marshal(&struct{
		Base int `json:"base"
		Times int `json:"times"`
	}{
		Base: 4,
		Times: 320,
	}); err != nil {
		errorf(err)
		return
	}

	if err = rt.Publish(msg); err != nil {
		errorf(err)
		return
	}

	// Now consume all messages till either an error or an
	// ok message come.
	for {
		msg, err = rt.Recv(ctx)
		if err != nil {
			errorf(err)
			return
		}
		switch msg.Type {
		case "ok":
			fmt.Printf("OK received!\n")
			return
		case "error":
			fmt.Printf("ERROR\n")
			return
		default:
			fmt.Printf("message %v received. Looping\n", msg.Type)
		}
	}
}
