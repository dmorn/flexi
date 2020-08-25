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
	"strconv"

	"github.com/jecoz/flexi"
	"github.com/jecoz/flexi/fargate"
	"github.com/Harvey-OS/ninep/protocol"
)

func main() {
	p := flag.Int("p", protocol.PORT, "Server listening port")
	b := flag.String("b", os.Args[0]+".backup", "Backup directory path containing information about each imported remote process")
	flag.Parse()

	addr := net.JoinHostPort("", strconv.Itoa(*p))
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
