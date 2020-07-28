package file

import (
	"errors"
	"io"
	"os"
	"sync"
	"time"
)

type Plumber struct {
	f    func(*Plumber) bool
	name string

	sync.Mutex
	buf     *LimitBuffer
	plumbed bool
	modTime time.Time
}

func (p *Plumber) Size() int64                       { return p.buf.Size() }
func (p *Plumber) Open() (io.ReadWriteCloser, error) { return p, nil }
func (p *Plumber) Stat() (os.FileInfo, error) {
	return Info{
		name:    p.name,
		size:    p.Size(),
		mode:    0222,
		modTime: p.modTime,
		isDir:   false,
	}, nil
}

func (p *Plumber) Read(b []byte) (int, error) {
	// Read is only called from the inside to obtain
	// buffer's contents, usually only after Close
	// is called (hence to writes occur).
	return p.buf.Read(b)
}

func (p *Plumber) Write(b []byte) (int, error) {
	p.Lock()
	defer p.Unlock()
	if p.plumbed {
		// We've plumbed successfully already.
		// Write is no longer allowed.
		return 0, errors.New("plumbed already")
	}
	return p.buf.Write(b)
}

func (p *Plumber) Close() error {
	p.Lock()
	defer p.Unlock()

	if p.plumbed {
		// Plumbed already, see Write.
		return errors.New("plumbed already")
	}
	if err := p.buf.Close(); err != nil {
		return err
	}

	// Plumb only if there is a plumbed function
	// & the buffer contains some data.
	if p.f == nil {
		return nil
	}
	if p.buf.Size() > 0 {
		p.plumbed = p.f(p)
	}
	return nil
}

func NewPlumber(name string, f func(*Plumber) bool) *Plumber {
	return &Plumber{name: name, f: f, buf: &LimitBuffer{}, modTime: time.Now()}
}
