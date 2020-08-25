package main

import (
	"net"
	"flag"
	"os"
	"os/signal"
	"log"

	"github.com/Harvey-OS/ninep/protocol"
	"github.com/Harvey-OS/ninep/pkg/debugfs"
	"github.com/jecoz/flexi/process"
)

func main() {
	port := flag.String("port", "9pfs", "Server listening port")
	debug := flag.Bool("d", true, "Toggle debugging")
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
	nineln, err := protocol.NewListener(func() protocol.NineServer {
		srv := &process.Srv{}
		if *debug {
			return &debugfs.DebugFileServer{srv}
		}
		return srv
	})
	log.Printf("*** server listening on %v", ln.Addr())
	if err := nineln.Serve(ln); err != nil {
		log.Printf("error * %v", err)
	}
}
