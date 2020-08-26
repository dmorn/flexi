package synthfs

import (
	"bytes"
	"os"
	"errors"
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
	b *Buffer
	r *bytes.Reader
}

func (f *BufferFile) Seek(o int64, w int) (int64, error) { return f.r.Seek(o, w) }
func (f *BufferFile) Read(p []byte) (int, error) { return f.r.Read(p) }
func (f *BufferFile) ReadAt(b []byte, o int64) (int, error) { return f.r.ReadAt(b, o) }

func (f *BufferFile) Stat() (os.FileInfo, error) {
	return FileStat{
		name: f.b.Name,
		size: f.r.Size(),
		mode: f.b.Mode,
		modTime: f.b.modTime,
		isDir: false,
	}, nil
}

// Note: Truncate and Write are to be implemented together to support
// RW files.

func (f *BufferFile) Truncate(n int64) error {
	return errors.New("file buffer: truncate not supported")
}

func (f *BufferFile) Write(p []byte) (int, error) {
	return 0, errors.New("file buffer: write not supported")
}

func (f *BufferFile) Close() error {
	// TODO: once write/truncate are implemented, this is the moment
	// were we *could* sync the updated filebuffer with the buffer,
	// if any change took place.
	return nil
}

// Buffer is an Opener implementation supporting data buffering. Its
// zero value is ready to be used, buf the Name and Mode fields should be
// filled before the first Open() call.
type Buffer struct {
	b bytes.Buffer
	modTime time.Time
	Name string
	Mode os.FileMode
}

func (b *Buffer) Open() (fs.File, error) {
	if b.modTime.IsZero() {
		b.modTime = time.Now()
	}
	return &BufferFile{
		r: bytes.NewReader(b.b.Bytes()),
		b: b,
	}, nil
}

func (b *Buffer) Write(p []byte) (int, error) { return b.b.Write(p) }
func (b *Buffer) Truncate(n int64) error {
	b.b.Truncate(int(n))
	return nil
}
