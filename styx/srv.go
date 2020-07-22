// SPDX-FileCopyrightText: 2020 jecoz
//
// SPDX-License-Identifier: BSD-3-Clause

package styx

import (
	"errors"
	"io"
	"log"
	"net"
	"os"
	"time"

	"aqwari.net/net/styx"
	"github.com/jecoz/flexi"
)

type Srv struct {
	Mntpt string
	Ln    net.Listener
	S     flexi.Spawner

	index int
	fs    *fs
}

func (s *Srv) Serve() error {
	echo := styx.HandlerFunc(func(s *styx.Session) {
		for s.Next() {
			log.Printf("%q %T %s", s.User, s.Request(), s.Request().Path())
		}
	})
	srv := &styx.Server{
		Handler:  styx.Stack(echo, s),
		ErrorLog: log.New(os.Stderr, "", log.LstdFlags),
	}
	return srv.Serve(s.Ln)
}

func (srv *Srv) Serve9P(s *styx.Session) {
	for s.Next() {
		srv.fs.serveRequest(s.Request())
	}
}

func (srv *Srv) SpawnProcess(rc io.ReadCloser) error {
	defer rc.Close()
	return errors.New("spawn process: not implemented yet")
}

// ServeFlexi creates a styx process serving 9p over ln. To gracefully shutdown
// the server, close the listener.
func ServeFlexi(ln net.Listener, mntpt string, s flexi.Spawner) error {
	srv := &Srv{Mntpt: mntpt, Ln: ln, S: s}
	ctl := NewInputBuffer("ctl", srv.SpawnProcess)
	root := &Dir{
		Name:    "",
		Perm:    os.ModePerm,
		ModTime: time.Now(),
		Ls: func() []File {
			union := append([]File{ctl}, DiskLs(mntpt)()...)
			return union
		},
	}
	srv.fs = &fs{Root: root}

	return srv.Serve()
}
