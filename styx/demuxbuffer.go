package styx

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"time"
)

type DemuxBuffer struct {
	Name    string
	Perm    os.FileMode
	ModTime time.Time

	// TODO: protect from concurrent access.
	bucket  []byte
	writers []io.Writer
}

func (b *DemuxBuffer) Len() int64 { return int64(len(b.bucket)) }

// Writes is called from the process and its bytes should reach
// the mounted files.
func (b *DemuxBuffer) Write(p []byte) (int, error) {
	buf := make([]byte, len(p))
	if n := copy(buf, p); n != len(p) {
		return n, io.ErrShortWrite
	}
	b.bucket = append(b.bucket, buf...)
	return io.MultiWriter(b.writers...).Write(p)
}

func (b *DemuxBuffer) Close() error { return nil }

func (b *DemuxBuffer) Open() (interface{}, error) {
	p := make([]byte, b.Len())
	n := copy(p, b.bucket)
	if int64(n) != b.Len() {
		return nil, io.ErrShortWrite
	}

	return &DemuxBufferio{
		buf:   bytes.NewBuffer(p),
		demux: b,
		onClose: func(d *DemuxBufferio) {
			writers := make([]io.Writer, 0, len(b.writers))
			for _, v := range b.writers {
				if v != d {
					writers = append(writers, v)
				}
			}
			b.writers = writers
		},
	}, nil
}

func (b *DemuxBuffer) Stat() (os.FileInfo, error) { return DemuxBufferInfo{b}, nil }

func (b *DemuxBuffer) Truncate(size int64) error {
	return fmt.Errorf("operation not allowed on demux buffer")
}

type DemuxBufferInfo struct {
	*DemuxBuffer
}

func (d DemuxBufferInfo) Name() string       { return d.DemuxBuffer.Name }
func (d DemuxBufferInfo) Size() int64        { return d.DemuxBuffer.Len() }
func (d DemuxBufferInfo) Mode() os.FileMode  { return d.DemuxBuffer.Perm }
func (d DemuxBufferInfo) ModTime() time.Time { return d.DemuxBuffer.ModTime }
func (d DemuxBufferInfo) IsDir() bool        { return false }
func (d DemuxBufferInfo) Sys() interface{} {
	bio, _ := d.DemuxBuffer.Open()
	return bio
}

type DemuxBufferio struct {
	demux   *DemuxBuffer
	buf     *bytes.Buffer
	onClose func(*DemuxBufferio)
}

func (b *DemuxBufferio) Write(p []byte) (int, error) { return b.buf.Write(p) }
func (b *DemuxBufferio) Read(p []byte) (int, error)  { return b.buf.Read(p) }
func (b *DemuxBufferio) Close() error {
	if onClose := b.onClose; onClose != nil {
		onClose(b)
	}
	return nil
}
