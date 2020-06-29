package styx

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type File interface {
	// Returns true if file is a directory. Usually when it is
	// Sys() returns a *Dir instance.
	IsDir() bool
	// Stat returns file information.
	Stat() (os.FileInfo, error)
	// In case of a directory, Sys should return a styx.Directory
	// implementation. In case of a regular file, it should at least
	// implement Read, Write, Close. Better with Seek too.
	// https://godoc.org/aqwari.net/net/styx#Topen.Ropen
	Sys() interface{}
}

func filename(f File) string {
	info, err := f.Stat()
	if err != nil {
		return ""
	}
	return info.Name()
}

func find(path []string, n File) File {
	if len(path) == 0 {
		return nil
	}
	name := path[len(path)-1]

	switch {
	case name != filename(n):
		return nil
	case len(path) == 1:
		return n
	case !n.IsDir():
		return nil
	default:
		// The name match and the node is
		// supposed to be a directory. Find
		// its children and search there!
		dir, ok := n.(*Dir)
		if !ok {
			return nil
		}
		ch := make(chan File)
		for _, v := range dir.Files {
			go func(n File) {
				ch <- find(path[1:], n)
			}(v)
		}
		results := make([]File, len(dir.Files))
		for i := range results {
			results[i] = <-ch
		}
		for _, v := range results {
			if v != nil {
				return v
			}
		}
		return nil
	}
}

// Fs allows to build arbitrary file system hierarchies.
type Fs struct {
	Root File
}

func (fs *Fs) Find(path string) (File, error) {
	path = filepath.Clean(path)
	if path == string(filepath.Separator) {
		path = ""
	}

	fields := strings.Split(path, string(filepath.Separator))
	file := find(fields, fs.Root)
	if file == nil {
		return nil, os.ErrNotExist
	}

	return file, nil
}

func (fs *Fs) Stat(path string) (os.FileInfo, error) {
	f, err := fs.Find(path)
	if err != nil {
		return nil, err
	}
	return f.Stat()
}

func (fs *Fs) open(path string) (interface{}, error) {
	file, err := fs.Find(path)
	if err != nil {
		return nil, err
	}
	return file.Sys(), nil
}

type Dir struct {
	Name  string
	Files []File
	Perm  os.FileMode
}

func NewDir(name string, perm os.FileMode, files ...File) *Dir {
	return &Dir{
		Name:  name,
		Files: files,
		Perm:  perm,
	}
}

func (d *Dir) IsDir() bool {
	return true
}

func (d *Dir) Stat() (os.FileInfo, error) {
	return &DirInfo{dir: d, t: time.Now()}, nil
}

func (d *Dir) Sys() interface{} {
	return &DirReader{dir: d}
}

type DirInfo struct {
	dir *Dir
	t   time.Time
}

func (d *DirInfo) Name() string {
	return d.dir.Name
}

func (d *DirInfo) Size() int64 {
	return 0
}

func (d *DirInfo) Mode() os.FileMode {
	return d.dir.Perm | os.ModeDir
}

func (d *DirInfo) ModTime() time.Time {
	return d.t
}

func (d *DirInfo) IsDir() bool {
	return true
}

func (d *DirInfo) Sys() interface{} {
	return d.dir.Sys()
}

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

type Buffer struct {
	io.Reader
	io.Writer
}

func NewMemFile(name string, perm os.FileMode) *MemFile {
	return &MemFile{
		perm: perm,
		name: name,
		buf:  new(bytes.Buffer),
		t:    time.Now(),
	}
}

type MemFile struct {
	perm os.FileMode
	name string
	err  error

	sync.Mutex
	t   time.Time
	buf *bytes.Buffer
}

func (f *MemFile) Stat() (os.FileInfo, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f, nil
}

func (f *MemFile) ModTime() time.Time {
	f.Lock()
	defer f.Unlock()
	return f.t
}

func (f *MemFile) Mode() os.FileMode {
	return f.perm
}

func (f *MemFile) Name() string {
	return f.name
}

func (f *MemFile) Size() int64 {
	f.Lock()
	defer f.Unlock()
	return int64(f.buf.Len())
}

func (f *MemFile) Sys() interface{} {
	f.Lock()
	defer f.Unlock()

	var b bytes.Buffer
	if _, err := io.Copy(&b, f.buf); err != nil {
		f.err = fmt.Errorf("copy into file buffer: %w", err)
		return nil
	}

	return &Buffer{
		Reader: &b,
		Writer: f,
	}
}

func (f *MemFile) Write(p []byte) (int, error) {
	f.Lock()
	f.t = time.Now()
	defer f.Unlock()

	return f.buf.Write(p)
}

func (f *MemFile) Read(p []byte) (int, error) {
	f.Lock()
	defer f.Unlock()

	return f.buf.Read(p)
}

func (f *MemFile) IsDir() bool {
	return false
}
