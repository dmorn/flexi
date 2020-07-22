package main

import (
	"flag"
	"log"
	"net"
	"os"
	"os/signal"

	"github.com/jecoz/flexi/fargate"
	"github.com/jecoz/flexi/styx"
)

func main() {
	port := flag.String("port", "9pfs", "Server listening port")
	mntpt := flag.String("m", "pmnt", "Remote processes mount point")
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

	if err := styx.ServeFlexi(ln, *mntpt, new(fargate.Fargate)); err != nil {
		log.Printf("flexi server error * %v", err)
	}
}
