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
	"path/filepath"

	"github.com/jecoz/flexi"
	"github.com/jecoz/flexi/fargate"
)

func main() {
	port := flag.String("port", "9pfs", "Server listening port")
	mtpt := flag.String("m", "pmnt", "Remote processes mount point")
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

	n := filepath.Join(*mtpt, "n")
	b := filepath.Join(*mtpt, "backup")
	s := &fargate.Fargate{BackupDir: b, Backup: true}
	if err := flexi.ServeFlexi(ln, n, s); err != nil {
		log.Printf("flexi server error * %v", err)
	}
}
