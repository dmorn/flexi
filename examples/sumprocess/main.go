package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/jecoz/flexi"
	"github.com/jecoz/flexi/styx"
)

type Params struct {
	From int `json:"from"`
	To   int `json:"to"`
}

type Status struct {
	Desc string `json:"desc"`
}

type Retv struct {
	Sum int `json:"sum"`
}

type Processor struct {
	p flexi.Process
}

func (po *Processor) run(p flexi.Process) error {
	var params Params
	if err := json.NewDecoder(po.p.Ctl()).Decode(&params); err != nil {
		return fmt.Errorf("unable to decode params: %w", err)
	}

	log.Printf("summing from %d to %d\n", params.From, params.To)

	if params.From > params.To {
		return fmt.Errorf("invalid input: From(%d) > To(%d)", params.From, params.To)
	}

	sum := 0
	senc := json.NewEncoder(po.p.Status())
	for i := params.From; i <= params.To; i++ {
		sum += i
		if err := senc.Encode(&Status{
			Desc: fmt.Sprintf("tmp sum: %d", sum),
		}); err != nil {
			return err
		}
	}
	return json.NewEncoder(po.p.Retv()).Encode(&Retv{
		Sum: sum,
	})
}

func (po *Processor) Run(p flexi.Process) {
	po.p = p
	if err := po.run(p); err != nil {
		log.Printf("processor error * %v", err)
		if err := json.NewEncoder(po.p.Err()).Encode(struct {
			Err string `json:"error"`
		}{
			Err: err.Error(),
		}); err != nil {
			log.Printf("error sending processor error * %v", err)
		}
	}
}

func main() {
	port := flag.String("port", "9pfs", "Server listening port")
	flag.Parse()

	p := styx.NewProcess()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
	go func() {
		s := <-sig
		log.Printf("%v <- signal received\n", s)
		p.Close()
	}()

	log.Printf("9p server listening on: %v", *port)
	if err := p.Serve(*port, new(Processor).Run); err != nil {
		log.Printf("server error * %v", err)
	}
}
