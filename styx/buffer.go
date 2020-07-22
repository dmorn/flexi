// SPDX-FileCopyrightText: 2020 jecoz
//
// SPDX-License-Identifier: BSD-3-Clause

package styx

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"aqwari.net/net/styx"
)

type Buffer struct {
	bytes.Buffer

	Name    string
	Perm    os.FileMode
	ModTime time.Time
}

func (b *Buffer) Close() error { return nil }
func (b *Buffer) Size() int64  { return int64(b.Len()) }
func (b *Buffer) Truncate(n int64) error {
	b.Buffer.Truncate(int(n))
	return nil
}
func (b *Buffer) Stat(opener Opener) os.FileInfo {
	return &finfo{
		name:    b.Name,
		mode:    b.Perm,
		modTime: b.ModTime,
		isDir:   false,
		s:       b,
		o:       opener,
	}
}

type InputBufferHook func(*InputBuffer) error

var _ File = &InputBuffer{}

// InputBuffer collects data till it is explicitly closed.
// When it is, calls the OnClose function.
type InputBuffer struct {
	Buffer
	OnClose InputBufferHook
}

func (ib *InputBuffer) OpenDir() (styx.Directory, error) {
	return nil, errors.New("input buffer: open dir: not a directory")
}

type hackedCloser struct {
	io.ReadWriteCloser
	close func() error
}

func (hc *hackedCloser) Close() error { return hc.close() }
func newHackedCloser(rwc io.ReadWriteCloser, close func() error) *hackedCloser {
	return &hackedCloser{rwc, close}
}

func (ib *InputBuffer) OpenFile() (io.ReadWriteCloser, error) {
	return newHackedCloser(ib, func() error {
		if err := ib.OnClose(ib); err != nil {
			return fmt.Errorf("input buffer: %w", err)
		}
		return nil
	}), nil
}

func (ib *InputBuffer) Stat() (os.FileInfo, error) {
	return ib.Buffer.Stat(ib), nil
}

func NewInputBuffer(name string, h InputBufferHook) *InputBuffer {
	return &InputBuffer{
		Buffer: Buffer{
			Name:    name,
			Perm:    0222,
			ModTime: time.Now(),
		},
		OnClose: h,
	}
}

type stream struct {
	closed bool
	buf    chan byte
}

// Clients read the buffer from the buffer, as if it was a regular file.
func (s *stream) Read(p []byte) (int, error) {
	n := 0
	for i := 0; i < len(p); i++ {
		b, ok := <-s.buf
		if !ok {
			return n, io.EOF
		}
		n = i + 1
		p[i] = b
	}
	return n, nil
}

func (s *stream) Close() error {
	if !s.closed {
		close(s.buf)
		s.closed = true
	}
	return nil
}
func (s *stream) Write(p []byte) (int, error) {
	n := 0
	for i, v := range p {
		select {
		case s.buf <- v:
			n = i + 1
		default:
			return n, io.ErrShortWrite
		}
	}
	return n, nil
}

type OutputBuffer struct {
	Buffer

	sync.RWMutex
	writers map[io.ReadWriteCloser]struct{}
	closed  bool
}

var _ File = &OutputBuffer{}

// Write will unlock stream readers returned with Open. Up to that
// point, they'll be waiting for some input to come.
func (ob *OutputBuffer) Write(p []byte) (int, error) {
	ob.RLock()
	defer ob.RUnlock()

	writers := make([]io.Writer, 0, len(ob.writers)+1)
	// First append the buffer itself: when clients connect after
	// the process closes the output buffer, we'll read stuff from there.
	writers = append(writers, &ob.Buffer)
	for k, _ := range ob.writers {
		writers = append(writers, k)
	}
	return io.MultiWriter(writers...).Write(p)
}

// Close closes the output streams. Subsequent Read calls on the opened
// streams will return io.EOF. Close stops if it is not capable of closing
// a stream, but removes the writers that have been successfully closed.
// Keeping on calling Close till nil is returned will eventually close
// every stream.
func (ob *OutputBuffer) Close() error {
	ob.Lock()
	defer func() {
		ob.closed = true
		ob.Unlock()
	}()

	writers := make(map[io.ReadWriteCloser]struct{})
	for k, v := range ob.writers {
		writers[k] = v
	}

	for k, _ := range writers {
		if err := k.Close(); err != nil {
			return fmt.Errorf("output buffer: %w", err)
		}
		delete(ob.writers, k)
	}
	return nil
}

func (ob *OutputBuffer) OpenDir() (styx.Directory, error) {
	return nil, errors.New("output buffer: open dir: not a directory")
}

type nopCloser struct{ io.ReadWriter }

func (c *nopCloser) Close() error                   { return nil }
func NopCloser(rw io.ReadWriter) io.ReadWriteCloser { return &nopCloser{rw} }

func (ob *OutputBuffer) openBuffer() (io.ReadWriteCloser, error) {
	buf := bytes.NewBuffer(ob.Buffer.Bytes())
	return NopCloser(buf), nil
}

const MaxStreamBuffer = 4096

func (ob *OutputBuffer) OpenFile() (io.ReadWriteCloser, error) {
	if ob.closed {
		return ob.openBuffer()
	}

	ob.Lock()
	// output may still come. Create the stream.
	s := &stream{buf: make(chan byte, MaxStreamBuffer)}
	ob.writers[s] = struct{}{}
	ob.Unlock()

	return newHackedCloser(s, func() error {
		ob.Lock()
		defer ob.Unlock()
		// First remove the pipe from the list of writers, or it
		// may make the MultiWriter fail. Close the pipe afterwards.
		delete(ob.writers, s)
		return s.Close()
	}), nil
}

func (ob *OutputBuffer) Stat() (os.FileInfo, error) {
	return ob.Buffer.Stat(ob), nil
}

func NewOutputBuffer(name string) *OutputBuffer {
	return &OutputBuffer{
		Buffer: Buffer{
			Name:    name,
			Perm:    0444,
			ModTime: time.Now(),
		},
		writers: make(map[io.ReadWriteCloser]struct{}),
	}
}
