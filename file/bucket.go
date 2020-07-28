package file

import (
	"bytes"
	"io"
	"os"
	"sync"
	"time"
)

type Bucket struct {
	name    string
	mode    os.FileMode
	modTime time.Time
	sync.Mutex
	buf *LimitBuffer
}

func (b *Bucket) Open() (io.ReadWriteCloser, error) {
	b.Lock()
	defer b.Unlock()

	return struct {
		io.Reader
		io.WriteCloser
	}{
		Reader:      bytes.NewBuffer(b.buf.Bytes()),
		WriteCloser: b,
	}, nil
}
func (b *Bucket) Stat() (os.FileInfo, error) {
	b.Lock()
	defer b.Unlock()
	return Info{
		name:    b.name,
		size:    b.buf.Size(),
		mode:    b.mode,
		modTime: b.modTime,
		isDir:   false,
	}, nil
}

func (b *Bucket) Write(p []byte) (int, error) {
	b.Lock()
	defer b.Unlock()
	b.modTime = time.Now()
	return b.buf.Write(p)
}

func (b *Bucket) Close() error { return b.buf.Close() }

func NewBucket(name string, mode os.FileMode, max int64) *Bucket {
	return &Bucket{name: name, mode: mode, modTime: time.Now(), buf: &LimitBuffer{Max: max}}
}
