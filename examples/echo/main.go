package main

import (
	"flag"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"strings"

	"github.com/jecoz/flexi"
	"github.com/jecoz/flexi/json"
	"github.com/jecoz/flexi/styx"
)

func main() {
	port := flag.String("port", "9pfs", "Server listening port")
	flag.Parse()

	addr := net.JoinHostPort("", *port)
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		log.Printf("error * %v", err)
		os.Exit(1)
	}

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
	go func() {
		s := <-sig
		log.Printf("%v <- signal received\n", s)
		ln.Close()
	}()

	if err := styx.ServeProcess(ln, flexi.ProcessorFunc(func(i *flexi.Stdio) {
		b := new(strings.Builder)
		if _, err := io.Copy(b, i.In); err != nil {
			panic("copy: " + err.Error())
		}

		h := &json.ProcessHelper{i}
		if err := h.Retv(b.String()); err != nil {
			panic("retv: " + err.Error())
		}
	})); err != nil {
		log.Printf("server error * %v", err)
	}
}
