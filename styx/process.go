package styx

import (
	"io"

	"aqwari.net/net/styx"
)

type Process struct {
	ctl    *File
	status *File
	retv   *File
	err    *File
}

func NewProcess() (*Process, error) {
	ctl := NewFile("ctl", 0222)
	status := NewFile("status", 0444)
	retv := NewFile("retv", 0444)
	err := NewFile("err", 0444)

	return &Process{
		ctl:    ctl,
		status: status,
		retv:   retv,
		err:    err,
	}, nil
}

func (p *Process) Dir() (*Dir, error) {
	return NewDir(0555, p.ctl, p.status, p.retv, p.err)
}

func (p *Process) Ctl() io.Reader {
	return p.ctl
}

func (p *Process) Status() io.Writer {
	return p.status
}

func (p *Process) Retv() io.Writer {
	return p.retv
}

func (p *Process) Err() io.Writer {
	return p.err
}

func (p *Process) serveReq(t styx.Request) error {
	dir, err := p.Dir()
	if err != nil {
		return err
	}
	info, err := dir.Stat(t.Path())
	if err != nil {
		return err
	}

	switch msg := t.(type) {
	case styx.Topen:
		msg.Ropen(info.Sys(), nil)
	case styx.Twalk:
		msg.Rwalk(info, nil)
	case styx.Tstat:
		msg.Rstat(info, nil)
	default:
	}
	return nil
}

func (p *Process) Serve9P(s *styx.Session) {
	for s.Next() {
		t := s.Request()
		if err := p.serveReq(t); err != nil {
			t.Rerror(err.Error())
		}
	}
}
