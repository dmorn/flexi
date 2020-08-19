// SPDX-FileCopyrightText: 2020 jecoz
//
// SPDX-License-Identifier: BSD-3-Clause

package styx

import (
	"log"
	"net"
	"os"

	"aqwari.net/net/styx"
	"github.com/jecoz/flexi/fs"
)

type Srv struct {
	Ln net.Listener
}

func (s *Srv) Serve(handlers ...styx.Handler) error {
	srv := &styx.Server{
		Handler: styx.Stack(handlers...),
	}
	return srv.Serve(s.Ln)
}

func Serve(ln net.Listener, fs fs.FS) error {
	srv := &Srv{ln}
	return srv.Serve(
		&LogHandler{Log: log.New(os.Stderr, "", log.LstdFlags)},
		&FSHandler{FS: fs},
	)
}
