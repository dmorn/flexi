package synthfs

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/jecoz/flexi/fs"
)

// TODO: DRY (see fs.FileStat)
type FileStat struct {
	name    string
	size    int64
	mode    os.FileMode
	modTime time.Time
	isDir   bool
}

func (i FileStat) Name() string       { return i.name }
func (i FileStat) Size() int64        { return i.size }
func (i FileStat) Mode() os.FileMode  { return i.mode }
func (i FileStat) ModTime() time.Time { return i.modTime }
func (i FileStat) IsDir() bool        { return i.isDir }
func (i FileStat) Sys() interface{}   { return nil }

type BufferFile struct {
	uAt time.Time
	b   *Buffer
	r   *bytes.Reader
}

func (f *BufferFile) Seek(o int64, w int) (int64, error) { return f.sync(o, w) }
func (f *BufferFile) Read(p []byte) (int, error) {
	// Each time we perform Truncate, Read or Write operations
	// on f.b.b we need to update the underlying datasource of
	// this bytes.Reader, as it may be reading from a slice that
	// as changed.
	if f.uAt.After(f.b.uAt) {
		if _, err := f.sync(f.Offset(), io.SeekStart); err != nil {
			return 0, err
		}
	}
	return f.r.Read(p)
}
func (f *BufferFile) ReadAt(b []byte, o int64) (int, error) {
	if f.uAt.After(f.b.uAt) {
		if _, err := f.sync(0, io.SeekStart); err != nil {
			return 0, err
		}
	}
	return f.r.ReadAt(b, o)
}

func (f *BufferFile) Stat() (os.FileInfo, error) {
	return FileStat{
		name:    f.b.Name,
		size:    f.r.Size(),
		mode:    f.b.Mode,
		modTime: f.uAt,
		isDir:   false,
	}, nil
}

func (f *BufferFile) Offset() int64 {
	return f.r.Size() - int64(f.r.Len())
}

func (f *BufferFile) sync(off int64, w int) (abs int64, err error) {
	f.r.Reset(f.b.b.Bytes())
	if abs, err = f.r.Seek(off, w); err != nil {
		return
	}
	f.uAt = time.Now()
	f.b.uAt = f.uAt
	return
}

func (f *BufferFile) Truncate(size int64) (err error) {
	f.b.b.Truncate(int(size))
	_, err = f.sync(size, io.SeekStart)
	return
}

func (f *BufferFile) Write(p []byte) (n int, err error) {
	if f.b.uAt.After(f.uAt) {
		err = fmt.Errorf("buffer contents changed since last read")
		return
	}
	off := f.Offset()
	if n, err = f.b.b.Write(p); err != nil {
		return
	}
	_, err = f.sync(off, io.SeekStart)
	return
}

func (f *BufferFile) Close() error {
	return nil
}

// Buffer is an Opener implementation supporting data buffering. Its
// zero value is ready to be used, buf the Name and Mode fields should be
// filled before the first Open() call.
type Buffer struct {
	b    bytes.Buffer
	uAt  time.Time
	Name string
	Mode os.FileMode
}

// Open returns an fs.File implementation that interacts with the underlying
// buffer. The file returned implements io.Write and Truncate too and can
// be used to edit b's contents.
func (b *Buffer) Open() (fs.File, error) {
	if b.uAt.IsZero() {
		b.uAt = time.Now()
	}
	return &BufferFile{
		r:   bytes.NewReader(b.b.Bytes()),
		b:   b,
		uAt: b.uAt,
	}, nil
}
