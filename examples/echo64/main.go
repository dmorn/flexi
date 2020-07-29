// SPDX-FileCopyrightText: 2020 jecoz
//
// SPDX-License-Identifier: BSD-3-Clause

package main

import (
	"bytes"
	"encoding/base64"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/signal"

	"github.com/jecoz/flexi"
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

	if err := flexi.ServeProcess(ln, flexi.ProcessorFunc(func(i *flexi.Stdio) {
		h := &flexi.JsonHelper{}

		b := new(bytes.Buffer)
		if _, err := io.Copy(b, i.In); err != nil {
			h.Err(i.Err, fmt.Errorf("copy: %w", err))
			return
		}
		encoded := base64.StdEncoding.EncodeToString(b.Bytes())

		if err := h.Encode(i.Retv, &struct {
			Original string `json:"original"`
			Base64   string `json:"base64"`
		}{
			Original: b.String(),
			Base64:   encoded,
		}); err != nil {
			h.Err(i.Err, fmt.Errorf("retv: %w", err))
			return
		}
	})); err != nil {
		log.Printf("server error * %v", err)
	}
}
