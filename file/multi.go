// SPDX-FileCopyrightText: 2020 jecoz
//
// SPDX-License-Identifier: BSD-3-Clause

package file

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"sync"
	"time"
)

// Multi is a fs.File implementation that broadcasts
// whatever is written to it to all its Open readers.
// Before Close is called, each Open call returns a stream
// reader that never returs io.EOF buf rather blocks.
// When Close is called, all readers are unlocked and will
// eventually reach the end of the stream. After this point,
// Open simply returns a buffer containing its contents.
// Initialize it with NewMulti.
type Multi struct {
	name string
	buf  *LimitBuffer

	sync.RWMutex
	modTime time.Time
	writers map[io.ReadWriteCloser]struct{}
	closed  bool
}

// Write will make Multi act as an io.MultiWriter on
// each Writer in m.writers.
func (m *Multi) Write(p []byte) (int, error) {
	m.Lock()
	m.modTime = time.Now()
	m.Unlock()

	m.RLock()
	writers := make([]io.Writer, 0, len(m.writers)+1)

	// First append the buffer itself: when clients connect after
	// the process closes the output buffer, we'll read stuff from there.
	writers = append(writers, m.buf)
	for k, _ := range m.writers {
		writers = append(writers, k)
	}
	m.RUnlock()

	return io.MultiWriter(writers...).Write(p)
}

// Close closes the output streams. Subsequent Read calls on the opened
// streams will return io.EOF. Close stops if it is not capable of closing
// a stream, but removes the writers that have been successfully closed.
// Keeping on calling Close till nil is returned will eventually close
// every stream.
func (m *Multi) Close() error {
	m.Lock()
	defer func() {
		m.closed = true
		m.Unlock()
	}()

	writers := make(map[io.ReadWriteCloser]struct{})
	for k, v := range m.writers {
		writers[k] = v
	}

	for k, _ := range writers {
		if err := k.Close(); err != nil {
			return fmt.Errorf("output buffer: %w", err)
		}
		delete(m.writers, k)
	}
	return nil
}

type hackedCloser struct {
	io.ReadWriteCloser
	altClose func() error
}

func (c *hackedCloser) Close() error { return c.altClose() }

func hackClose(rwc io.ReadWriteCloser, f func() error) *hackedCloser {
	return &hackedCloser{rwc, f}
}

func (m *Multi) Open() (io.ReadWriteCloser, error) {
	m.Lock()
	defer m.Unlock()
	if m.closed {
		return &LimitBuffer{
			buf: *bytes.NewBuffer(m.buf.Bytes()),
		}, nil
	}

	// output may still come. Create the stream.
	p := newPipe()
	m.writers[p] = struct{}{}

	return hackClose(p, func() error {
		m.Lock()
		defer m.Unlock()
		// First remove the pipe from the list of writers, or it
		// may make the MultiWriter fail. Close the pipe afterwards.
		delete(m.writers, p)
		return p.Close()
	}), nil
}

func (m *Multi) Stat() (os.FileInfo, error) {
	m.RLock()
	defer m.RUnlock()

	mtime := m.modTime
	if mtime.IsZero() {
		mtime = time.Now()
	}

	return Info{
		name:    m.name,
		size:    m.buf.Size(),
		mode:    0444,
		modTime: mtime,
		isDir:   false,
	}, nil
}

func NewMulti(name string) *Multi {
	return &Multi{
		name:    name,
		writers: make(map[io.ReadWriteCloser]struct{}),
		buf:     &LimitBuffer{},
		modTime: time.Now(),
	}
}

// pipe is an in-memory pipe. Pretty much just a wrapper around
// an io.Pipe() call.
type pipe struct {
	r *io.PipeReader
	w *io.PipeWriter
}

func (p *pipe) Write(b []byte) (int, error) { return p.w.Write(b) }
func (p *pipe) Read(b []byte) (int, error)  { return p.r.Read(b) }
func (p *pipe) Close() error                { return p.w.Close() }

func newPipe() (p *pipe) {
	p = new(pipe)
	p.r, p.w = io.Pipe()
	return
}
