package styx

import (
	"io"
	"os"
	"time"
	"path/filepath"
)

type File interface {
	// Returns true if file is a directory. Usually when it is
	// Sys() returns a *DirReader instance.
	IsDir() bool
	// Stat returns file information.
	Stat() (os.FileInfo, error)
	// In case of a directory, Sys should return a styx.Directory
	// implementation. In case of a regular file, it should at least
	// implement Read, Write, Close. Better with Seek too.
	// https://godoc.org/aqwari.net/net/styx#Topen.Ropen
	Sys() interface{}
}

type Fs struct {
	table map[string]File
}

func NewFs() *Fs { return &Fs{ table: make(map[string]File) } }

func (fs *Fs) Add(path string, f File) error {
	if _, ok := fs.table[path]; ok {
		return os.ErrExist
	}
	fs.table[path] = f
	return nil
}

func (fs *Fs) Lookup(path string) (File, error) {
	f, ok := fs.table[path]
	if !ok {
		return nil, os.ErrNotExist
	}
	return f, nil
}

func (fs *Fs) Stat(path string) (os.FileInfo, error) {
	f, err := fs.Lookup(path)
	if err != nil {
		return nil, err
	}
	return f.Stat()
}

func (fs *Fs) Open(path string) (interface{}, error) {
	f, err := fs.Lookup(path)
	if err != nil {
		return nil, err
	}
	return f.Sys(), nil
}

type Dir struct {
	path  string
	Files []File
	Perm  os.FileMode
}

func NewDir(path string, perm os.FileMode, files ...File) *Dir {
	return &Dir{
		path:  path,
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
	return filepath.Dir(d.dir.path)
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

type MemFile struct {
	perm os.FileMode
	path string
	err  error

	r *io.PipeReader
	w *io.PipeWriter
}

func NewMemFile(path string, perm os.FileMode) *MemFile {
	r, w := io.Pipe()
	return &MemFile{
		perm: perm,
		path: path,
		r: r,
		w: w,
	}
}


func (f *MemFile) Stat() (os.FileInfo, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f, nil
}

func (f *MemFile) ModTime() time.Time {
	return time.Now()
}

func (f *MemFile) Mode() os.FileMode {
	return f.perm
}

func (f *MemFile) Name() (name string) {
	_, name = filepath.Split(f.path)
	return
}

func (f *MemFile) Size() int64 {
	return 0
}

func (f *MemFile) Sys() interface{} {
	return f
}

func (f *MemFile) Write(p []byte) (int, error) {
	return f.w.Write(p)
}

func (f *MemFile) Read(p []byte) (int, error) {
	return f.r.Read(p)
}

func (f *MemFile) IsDir() bool {
	return false
}
