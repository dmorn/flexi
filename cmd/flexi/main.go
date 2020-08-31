// SPDX-FileCopyrightText: 2020 jecoz
//
// SPDX-License-Identifier: BSD-3-Clause

package main

import (
	"flag"
	"log"
	"net"
	"os"
	"os/signal"

	"github.com/jecoz/flexi"
	"github.com/jecoz/flexi/fargate"
)

func main() {
	port := flag.String("p", "9pfs", "Server listening port")
	bkup := flag.String("b", os.Args[0]+".bkup", "Process backup path. Used for recoverying itermediate states")
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

	s := &fargate.Fargate{BackupDir: *bkup, Backup: true}
	if err := flexi.Serve(ln, s); err != nil {
		log.Printf("flexi server error * %v", err)
	}
}
