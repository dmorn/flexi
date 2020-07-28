// SPDX-FileCopyrightText: 2020 jecoz
//
// SPDX-License-Identifier: BSD-3-Clause

package styx

import (
	"log"
	"net"

	"aqwari.net/net/styx"
	"github.com/jecoz/flexi/file"
	"github.com/jecoz/flexi/fs"
)

type Srv struct {
	Ln net.Listener
	FS fs.FS
}

func (srv *Srv) handleRequest(t styx.Request) {
	switch msg := t.(type) {
	case styx.Ttruncate:
		// TODO: implement if needed
		msg.Rtruncate(nil)
	case styx.Tutimes:
		// TODO: implement if needed
		msg.Rutimes(nil)
	case styx.Tcreate:
		b := file.NewBucket(msg.Name, msg.Mode, 2048)
		if err := srv.FS.Create(msg.Path(), b); err != nil {
			msg.Rerror(err.Error())
			return
		}
		msg.Rcreate(b.Open())
	case styx.Topen, styx.Twalk, styx.Tstat:
		// All these messages require an open first.
		// We're taking care of it in a single place.
	default:
		// Handled with default responses.
		return
	}

	file, err := srv.FS.Open(t.Path())
	if err != nil {
		t.Rerror(err.Error())
		return
	}
	switch msg := t.(type) {
	case styx.Topen:
		msg.Ropen(file.Open())
	case styx.Twalk:
		msg.Rwalk(file.Stat())
	case styx.Tstat:
		msg.Rstat(file.Stat())
	}
}

func (srv *Srv) Serve9P(s *styx.Session) {
	for s.Next() {
		srv.handleRequest(s.Request())
	}
}

func (s *Srv) Serve() error {
	echo := styx.HandlerFunc(func(s *styx.Session) {
		for s.Next() {
			log.Printf("%q %T %s", s.User, s.Request(), s.Request().Path())
		}
	})
	srv := &styx.Server{
		Handler: styx.Stack(echo, s),
	}
	return srv.Serve(s.Ln)
}

func Serve(ln net.Listener, fs fs.FS) error {
	srv := &Srv{Ln: ln, FS: fs}
	return srv.Serve()
}
