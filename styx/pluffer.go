// SPDX-FileCopyrightText: 2020 jecoz
//
// SPDX-License-Identifier: BSD-3-Clause

package styx

import (
	"bytes"
	"os"
	"time"
)

// Pluffer acts like a plumbing buffer (in plan9 terms). Upon close, the
// Pluffer calls OnClose function letting callers act on the stored bytes.
type Pluffer struct {
	Name    string
	Perm    os.FileMode
	ModTime time.Time
	OnClose func(*Pluffer) error

	buf bytes.Buffer
}

func (p *Pluffer) Stat() (os.FileInfo, error) { return plufferInfo{p}, nil }
func (p *Pluffer) Open() (interface{}, error) { return &plufferio{p}, nil }
func (p *Pluffer) Truncate(size int64) error {
	p.buf.Truncate(int(size))
	return nil
}

func (p *Pluffer) Len() int64                  { return int64(p.buf.Len()) }
func (pf *Pluffer) Read(p []byte) (int, error) { return pf.buf.Read(p) }

type plufferInfo struct {
	*Pluffer
}

func (p plufferInfo) Name() string       { return p.Pluffer.Name }
func (p plufferInfo) Size() int64        { return int64(p.Pluffer.buf.Len()) }
func (p plufferInfo) Mode() os.FileMode  { return p.Pluffer.Perm }
func (p plufferInfo) ModTime() time.Time { return p.Pluffer.ModTime }
func (p plufferInfo) IsDir() bool        { return false }
func (p plufferInfo) Sys() interface{} {
	rwc, _ := p.Pluffer.Open()
	return rwc
}

type plufferio struct {
	pluffer *Pluffer
}

func (pf *plufferio) Read(p []byte) (int, error)  { return pf.pluffer.buf.Read(p) }
func (pf *plufferio) Write(p []byte) (int, error) { return pf.pluffer.buf.Write(p) }
func (pf *plufferio) Close() error {
	if onClose := pf.pluffer.OnClose; onClose != nil {
		return onClose(pf.pluffer)
	}
	return nil
}
