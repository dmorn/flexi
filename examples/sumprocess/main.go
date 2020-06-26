package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/jecoz/flexi"
	"github.com/jecoz/flexi/ninep"
)

type Params struct {
	From int `json:"from"`
	To   int `json:"to"`
}

type Processor struct {
	p flexi.Process
}

func (po *Processor) Params() (p Params, err error) {
	err = json.NewDecoder(po.p.Ctl()).Decode(&p)
	return
}

func (po *Processor) Status(sum int) error {
	return json.NewEncoder(po.p.Status()).Encode(struct {
		D string `json:"description"`
	}{
		D: fmt.Sprintf("tmp sum: %d", sum),
	})
}

func (po *Processor) Retv(sum int) error {
	return json.NewEncoder(po.p.Retv()).Encode(struct {
		Sum int `json:"sum"`
	}{
		Sum: sum,
	})
}

func (po *Processor) run(p flexi.Process) error {
	params, err := po.Params()
	if err != nil {
		return err
	}

	log.Printf("summing from %d to %d\n", params.From, params.To)

	if params.From > params.To {
		return fmt.Errorf("invalid input: From(%d) > To(%d)", params.From, params.To)
	}

	sum := 0
	for i := params.From; i <= params.To; i++ {
		sum += i
		if err := po.Status(sum); err != nil {
			return err
		}
	}
	return po.Retv(sum)
}

func (po *Processor) Run(p flexi.Process) {
	po.p = p
	if err := po.run(p); err != nil {
		po.Err(err)
	}
}

func (po *Processor) Err(err error) {
	log.Printf("processor error * %v", err)
	if err := json.NewEncoder(po.p.Err()).Encode(struct {
		Err string `json:"error"`
	}{
		Err: err.Error(),
	}); err != nil {
		log.Printf("error sending processor error * %v", err)
	}
}

func main() {
	srv := &ninep.Srv{Port: "9pfs"}

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
	go func() {
		s := <-sig
		log.Printf("%v <- signal received\n", s)
		srv.Close()
	}()

	if err := srv.Run(new(Processor).Run); err != nil {
		log.Printf("server error * %v", err)
	}
}
