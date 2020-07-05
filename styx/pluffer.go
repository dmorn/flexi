package styx

import (
	"os"
	"bytes"
	"time"
)

type Pluffer struct {
	Name string
	Perm os.FileMode
	ModTime time.Time
	OnClose func(*Pluffer) error

	buf bytes.Buffer
}

func (p *Pluffer) Stat() (os.FileInfo, error) { return PlufferInfo{p}, nil }
func (p *Pluffer) Open() (interface{}, error) { return &Plufferio{p}, nil }
func (p *Pluffer) Truncate(size int64) error {
	p.buf.Truncate(int(size))
	return nil
}

func (p *Pluffer) Len() int64 { return int64(p.buf.Len()) }
func (pf *Pluffer) Read(p []byte) (int, error) { return pf.buf.Read(p) }

type PlufferInfo struct {
	*Pluffer
}

func (p PlufferInfo) Name() string       { return p.Pluffer.Name }
func (p PlufferInfo) Size() int64        { return int64(p.Pluffer.buf.Len()) }
func (p PlufferInfo) Mode() os.FileMode  { return p.Pluffer.Perm }
func (p PlufferInfo) ModTime() time.Time { return p.Pluffer.ModTime }
func (p PlufferInfo) IsDir() bool        { return false }
func (p PlufferInfo) Sys() interface{}   {
	rwc, _ := p.Pluffer.Open()
	return rwc
}

type Plufferio struct {
	pluffer *Pluffer
}

func (pf *Plufferio) Read(p []byte) (int, error) { return pf.pluffer.buf.Read(p) }
func (pf *Plufferio) Write(p []byte) (int, error) { return pf.pluffer.buf.Write(p) }
func (pf *Plufferio) Close() error {
	if onClose := pf.pluffer.OnClose; onClose != nil {
		return onClose(pf.pluffer)
	}
	return nil
}
