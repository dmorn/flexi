package styx

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

type File interface {
	// Stat returns file information.
	Stat() (os.FileInfo, error)
	// In case of a directory, Sys should return a styx.Directory
	// implementation. In case of a regular file, it should at least
	// implement Read, Write, Close. Better with Seek too.
	// https://godoc.org/aqwari.net/net/styx#Topen.Ropen
	Sys() interface{}
}

type RelayBuffer struct {
	r       io.Reader
	w       *io.PipeWriter
	OnClose func() error
}

func (b *RelayBuffer) Close() error {
	if onClose := b.OnClose; onClose != nil {
		return onClose()
	}
	return nil
}

func (b *RelayBuffer) Write(p []byte) (int, error) { return b.w.Write(p) }
func (b *RelayBuffer) Read(p []byte) (int, error)  { return b.r.Read(p) }

type ProxyBuffer struct {
	r       io.Reader
	w       io.Writer
	bucket  []byte
	readers []io.Reader
	writers []io.Writer
}

func (b *ProxyBuffer) Read(p []byte) (int, error) {
	if b.r == nil {
		return 0, io.EOF
	}
	return b.r.Read(p)
}

func (b *ProxyBuffer) Write(p []byte) (int, error) {
	buf := make([]byte, len(p))
	n := copy(buf, p)
	if n != len(p) {
		return 0, io.ErrShortWrite
	}
	b.bucket = append(b.bucket, buf...)
	if b.w == nil {
		return n, nil
	}

	return b.w.Write(p)
}

func (b *ProxyBuffer) NewRelay() *RelayBuffer {
	// Relay writes to a blocking pipe that needs a
	// process reading on the other side.
	piper, pipew := io.Pipe()
	if b.r == nil {
		b.r = io.MultiReader()
	}
	b.r = io.MultiReader(b.r, piper)

	// Relay reads from a buffer that gets filled
	// with all data produced by the process.
	bucket := make([]byte, len(b.bucket))
	if n := copy(bucket, b.bucket); n != len(b.bucket) {
		panic("new relay: short copy")
	}
	buf := bytes.NewBuffer(bucket)

	if b.writers == nil {
		b.writers = make([]io.Writer, 0, 10)
	}
	b.writers = append(b.writers, buf)
	if b.w == nil {
		b.w = io.MultiWriter()
	}
	b.w = io.MultiWriter(b.w, buf)

	if b.readers == nil {
		b.readers = make([]io.Reader, 0, 10)
	}
	if b.r == nil {
		b.r = io.MultiReader()
	}
	b.r = io.MultiReader(b.r, piper)

	return &RelayBuffer{
		OnClose: func() error {
			readers := make([]io.Reader, 0, len(b.readers))
			for _, v := range b.readers {
				if v != piper {
					readers = append(readers, v)
				}
			}
			if len(readers) == len(b.readers) {
				return fmt.Errorf("close called multiple times on buffer")
			}

			writers := make([]io.Writer, 0, len(b.writers))
			for _, v := range b.writers {
				if v != buf {
					writers = append(writers, v)
				}
			}
			if len(writers) == len(b.writers) {
				return fmt.Errorf("close called multiple times on buffer")
			}

			pipew.Close()
			buf.Reset()
			buf = nil
			b.writers = writers
			b.readers = readers
			b.r = io.MultiReader(readers...)
			b.w = io.MultiWriter(writers...)
			return nil
		},
		w: pipew,
		r: buf,
	}
}

type VolFile struct {
	ProxyBuffer
	Perm os.FileMode
	Name string
}

func (f *VolFile) Stat() (os.FileInfo, error) { return &VolFileInfo{f}, nil }
func (f *VolFile) Sys() interface{}           { return f.NewRelay() }

type VolFileInfo struct {
	f *VolFile
}

func (i *VolFileInfo) Name() string       { return i.f.Name }
func (i *VolFileInfo) Size() int64        { return 0 }
func (i *VolFileInfo) Mode() os.FileMode  { return i.f.Perm }
func (i *VolFileInfo) ModTime() time.Time { return time.Now() }
func (i *VolFileInfo) IsDir() bool        { return false }
func (i *VolFileInfo) Sys() interface{}   { return i.f.NewRelay() }

type Dir struct {
	Name  string
	Files []File
	Perm  os.FileMode
}

func (d *Dir) Stat() (os.FileInfo, error) {
	return &DirInfo{dir: d, t: time.Now()}, nil
}

func (d *Dir) Sys() interface{} { return &DirReader{dir: d} }

type DirInfo struct {
	dir *Dir
	t   time.Time
}

func (d *DirInfo) Name() string       { return filepath.Dir(d.dir.Name) }
func (d *DirInfo) Size() int64        { return 0 }
func (d *DirInfo) Mode() os.FileMode  { return d.dir.Perm | os.ModeDir }
func (d *DirInfo) ModTime() time.Time { return d.t }
func (d *DirInfo) IsDir() bool        { return true }
func (d *DirInfo) Sys() interface{}   { return d.dir.Sys() }

type DirReader struct {
	offset int
	dir    *Dir
}

func (d *DirReader) Readdir(n int) ([]os.FileInfo, error) {
	if d.dir == nil {
		return nil, os.ErrInvalid
	}
	if d.offset >= len(d.dir.Files) {
		return nil, io.EOF
	}
	files := d.dir.Files[d.offset:]
	count := len(files)
	take := n
	if take <= 0 || take > count {
		take = count
	}

	files = files[:take]
	fis := make([]os.FileInfo, len(files))
	for i, v := range files {
		info, err := v.Stat()
		if err != nil {
			return nil, err
		}
		fis[i] = info
	}
	d.offset += take
	return fis, nil
}
