// SPDX-FileCopyrightText: 2020 jecoz
//
// SPDX-License-Identifier: BSD-3-Clause

package styx

import (
	"log"

	"aqwari.net/net/styx"
	"github.com/jecoz/flexi/file"
	"github.com/jecoz/flexi/fs"
)

type LogHandler struct {
	Log *log.Logger
}

func (h *LogHandler) Serve9P(s *styx.Session) {
	for s.Next() {
		h.Log.Printf("%q %T %s", s.User, s.Request(), s.Request().Path())
	}
}

type FSHandler struct {
	FS fs.FS
}

func (h *FSHandler) handleRequest(t styx.Request) {
	switch msg := t.(type) {
	case styx.Tremove:
		msg.Rremove(h.FS.Remove(msg.Path()))
		return
	case styx.Ttruncate:
		// TODO: implement if needed
		msg.Rtruncate(nil)
		return
	case styx.Tutimes:
		// TODO: implement if needed
		msg.Rutimes(nil)
		return
	case styx.Tcreate:
		b := file.NewBucket(msg.Name, msg.Mode, 2048)
		if err := h.FS.Create(msg.Path(), b); err != nil {
			msg.Rerror(err.Error())
			return
		}
		msg.Rcreate(b.Open())
		return
	case styx.Topen, styx.Twalk, styx.Tstat:
		// All these messages require an open first.
		// We're taking care of it in a single place.
	default:
		// Handled with default responses.
		return
	}

	file, err := h.FS.Open(t.Path())
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

func (h *FSHandler) Serve9P(s *styx.Session) {
	for s.Next() {
		h.handleRequest(s.Request())
	}
}
