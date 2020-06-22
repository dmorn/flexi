package rabbit

import (
	"context"
	"fmt"
	"time"

	"github.com/jecoz/flexi"
	"github.com/streadway/amqp"
)

func Dial(addr string) (*amqp.Connection, error) {
	return amqp.Dial(addr)
}

type Transport struct {
	queue string
	conn  *amqp.Connection
	ch    *amqp.Channel
}

func NewTransport(url string, queue string) (*Transport, error) {
	conn, err := Dial(url)
	if err != nil {
		return nil, err
	}
	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	durable := true
	autodel := false
	if _, err = ch.QueueDeclare(queue, durable, autodel, false, false, nil); err != nil {
		return nil, err
	}

	return &Transport{
		queue: queue,
		conn:  conn,
		ch:    ch,
	}, nil
}

// The queue this tranport will try to connect to must be already present.
func NewTransportPassive(url string, queue string) (*Transport, error) {
	conn, err := Dial(url)
	if err != nil {
		return nil, err
	}
	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	durable := false
	autodel := true
	if _, err = ch.QueueDeclarePassive(queue, durable, autodel, false, false, nil); err != nil {
		return nil, err
	}

	return &Transport{
		queue: queue,
		conn:  conn,
		ch:    ch,
	}, nil
}

func (t *Transport) Close() error {
	if err := t.ch.Close(); err != nil {
		return err
	}
	return t.conn.Close()
}

func (t *Transport) Recv(ctx context.Context) (*flexi.Msg, error) {
	ch, err := t.conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("unable to open recv channel: %w", err)
	}
	defer ch.Close()

	msgrx := make(chan amqp.Delivery, 1)
	errrx := make(chan error, 1)
	go func() {
		defer close(msgrx)
		defer close(errrx)
		msg, _, err := ch.Get(t.queue, true)
		if err != nil {
			errrx <- fmt.Errorf("getting response: %w", err)
			return
		}
		msgrx <- msg
	}()

	select {
	case <-ctx.Done():
		ch.Close()
		<-errrx
		<-msgrx
		return nil, fmt.Errorf("recv invalidated: %w", ctx.Err())
	case msg := <-msgrx:
		return &flexi.Msg{
			ContentType: msg.ContentType,
			Type:        msg.Type,
			Body:        msg.Body,
		}, nil
	case err := <-errrx:
		return nil, err
	}
}

func (t *Transport) Publish(msg *flexi.Msg) error {
	return t.ch.Publish("", t.queue, true, true, amqp.Publishing{
		DeliveryMode: amqp.Transient,
		Timestamp:    time.Now(),
		ContentType:  msg.ContentType,
		Type:         msg.Type,
		Body:         msg.Body,
	})
}
