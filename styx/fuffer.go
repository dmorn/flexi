// SPDX-FileCopyrightText: 2020 jecoz
//
// SPDX-License-Identifier: BSD-3-Clause

package styx

import (
	"bytes"
	"errors"
	"io"
	"os"
	"time"

	"aqwari.net/net/styx"
)

type nopCloser struct {
	io.ReadWriter
}

func (c *nopCloser) Close() error { return nil }

func NopCloser(rw io.ReadWriter) io.ReadWriteCloser {
	return &nopCloser{rw}
}

// Fuffer is a File implemetation of a bytes.Buffer.
type Fuffer struct {
	Name    string
	Perm    os.FileMode
	ModTime time.Time

	Open func() (io.ReadWriteCloser, error)
	b    bytes.Buffer
}

func (f *Fuffer) OpenFile() (io.ReadWriteCloser, error) {
	if f.Open == nil {
		return NopCloser(&f.b), nil
	}
	return f.Open()
}

func (f *Fuffer) OpenDir() (styx.Directory, error) {
	return nil, errors.New("fuffer: open dir: not a directory")
}

func (f *Fuffer) Stat() (os.FileInfo, error) {
	return &finfo{
		name:    f.Name,
		mode:    f.Perm,
		modTime: f.ModTime,
		isDir:   false,
		s:       f,
		o:       f,
	}, nil
}

func (f *Fuffer) Close() error                { return nil }
func (f *Fuffer) Write(p []byte) (int, error) { return f.b.Write(p) }
func (f *Fuffer) Read(p []byte) (int, error)  { return f.b.Read(p) }
func (f *Fuffer) Size() int64                 { return int64(f.b.Len()) }
func (f *Fuffer) Truncate(n int64) error {
	f.b.Truncate(int(n))
	return nil
}

type FufferHook func(*Fuffer) error

type closeFuffer struct {
	f *Fuffer
	h FufferHook
}

func (c *closeFuffer) Close() error                { return c.h(c.f) }
func (c *closeFuffer) Read(p []byte) (int, error)  { return c.f.Read(p) }
func (c *closeFuffer) Write(p []byte) (int, error) { return c.f.Write(p) }

// OnClose will execute h each time a rwc returned from OpenFile gets
// closed. Calling this function will set the Open field.
func (f *Fuffer) OnClose(h FufferHook) {
	f.Open = func() (io.ReadWriteCloser, error) {
		return &closeFuffer{f: f, h: h}, nil
	}
}

type pipeFuffer struct {
}

func NewFuffer(name string, perm os.FileMode) *Fuffer {
	return &Fuffer{Name: name, Perm: perm}
}

func NewCloseFuffer(name string, perm os.FileMode, h FufferHook) (f *Fuffer) {
	f = NewFuffer(name, perm)
	f.OnClose(h)
	return
}

type errWriter struct {
	io.Reader
}

func (w *errWriter) Write(p []byte) (int, error) {
	return 0, errors.New("write not allowed")
}

func ErrWriter(r io.Reader) io.ReadWriter { return &errWriter{r} }

func NewPipeFuffer(name string, perm os.FileMode) (f *Fuffer) {
	f = NewFuffer(name, perm)
	f.Open = func() (io.ReadWriteCloser, error) {
		r := bytes.NewReader(f.b.Bytes())
		rw := ErrWriter(r)
		rwc := NopCloser(rw)
		return rwc, nil
	}
	return
}
