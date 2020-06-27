package styx

import (
	"bytes"
	"fmt"
	"os"
	"sync"
	"time"
)

type File struct {
	mode os.FileMode
	name string
	buf  *bytes.Buffer

	sync.Mutex
	t time.Time
}

func (f *File) String() string {
	return f.name
}

func NewFile(name string, mode os.FileMode) *File {
	return &File{
		mode: mode,
		name: name,
		buf:  new(bytes.Buffer),
		t:    time.Now(),
	}
}

func (f *File) IsValid() error {
	switch {
	case f.name == "":
		return fmt.Errorf("missing filename")
	default:
		return nil
	}
}

// Will be used by styx for Topen requests.
// See https://pkg.go.dev/aqwari.net/net/styx?tab=doc#Topen.Ropen
// func (f *File) Stat() (os.FileInfo, error) {
// 	return f, nil
// }

func (f *File) Write(p []byte) (int, error) {
	f.Lock()
	f.t = time.Now()
	f.Unlock()

	return f.buf.Write(p)
}

func (f *File) Read(p []byte) (int, error) {
	return f.buf.Read(p)
}

func (f *File) Name() string {
	return f.name
}

func (f *File) Size() int64 {
	return int64(f.buf.Len())
}

func (f *File) Mode() os.FileMode {
	return f.mode
}

func (f *File) ModTime() time.Time {
	f.Lock()
	defer f.Unlock()
	return f.t
}

func (f *File) IsDir() bool {
	return f.Mode().IsDir()
}

func (f *File) Sys() interface{} {
	return f.buf
}
