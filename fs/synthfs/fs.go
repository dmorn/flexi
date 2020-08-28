package synthfs

import (
	"errors"
	"fmt"
	"os"
	"path"
	"sort"
	"strings"
	"time"

	"github.com/jecoz/flexi/fs"
)

func PathFields(p string) (f []string) {
	var b string
	for {
		p = strings.TrimSuffix(p, "/")
		b = path.Base(p)
		if b == "." {
			sort.Reverse(sort.StringSlice(f))
			return
		}
		f = append(f, b)
		p = strings.TrimSuffix(p, b)
	}
}

// Opener describes the Open function, used to see some content
// as an open File.
type Opener interface {
	Open() (fs.File, error)
}

type OpenerFunc func() (fs.File, error)

func (o OpenerFunc) Open() (fs.File, error) { return o() }

type onode struct {
	Opener
	Name   string
	Parent *onode
	Leafs  []*onode
}

func (n *onode) traverse(p ...string) *onode {
	if len(p) == 0 {
		return n
	}
	h, t := p[0], p[1:]
	for _, v := range n.Leafs {
		if v.Name == h {
			return v.traverse(t...)
		}
	}
	return nil
}

func (n *onode) dirFileOpener() Opener {
	modTime := time.Now()
	return OpenerFunc(func() (fs.File, error) {
		fi := []os.FileInfo{}
		for _, v := range n.Leafs {
			f, err := v.Open()
			if err != nil {
				// NOTE: we could also return with what we have.
				return nil, err
			}
			s, err := f.Stat()
			if err != nil {
				// NOTE: same as above.
				return nil, err
			}
			fi = append(fi, s)
		}
		d := &fs.DirFile{
			ModTime: modTime,
			Name:    n.Name,
			Files:   fi,
		}
		return d, nil
	})
}

type FS struct {
	tree *onode
}

func (fsys *FS) InsertOpener(o Opener, n string, in ...string) error {
	if fsys.tree == nil {
		// TODO: what about the opener? It should be a directory.
		root := new(onode)
		root.Opener = root.dirFileOpener()
		fsys.tree = root
	}
	dirpath := path.Join(in...)
	var pn *onode
	if pn = fsys.tree.traverse(in...); pn == nil {
		return fmt.Errorf("traverse %v: %w", dirpath, fs.ErrNotExist)
	}

	// The parent node needs to be a directory, otherwise we end up
	// creating things that do not have sense.
	f, err := pn.Open()
	if err != nil {
		return fmt.Errorf("open parent: %w", err)
	}
	if !fs.FileIsDir(f) {
		return fmt.Errorf("cannot add %v to %v: not a directory", n, dirpath)
	}

	// Before adding the leaf, check that a file named
	// n is not already there.
	if ln := pn.traverse(n); ln != nil {
		return fs.ErrExist
	}
	if len(pn.Leafs) == 0 {
		pn.Leafs = []*onode{}
	}
	pn.Leafs = append(pn.Leafs, &onode{
		Opener: o,
		Name:   n,
		// No need to initialize Leafs now, we can postpone till
		// we actually want to add an opener to this onode too.
		Parent: pn,
	})
	return nil
}

func (fsys *FS) Open(p string) (fs.File, error) {
	if fsys.tree == nil {
		return nil, fs.ErrNotExist
	}
	n := fsys.tree.traverse(PathFields(p)...)
	if n == nil {
		return nil, fs.ErrNotExist
	}
	return n.Open()
}

// Create adds a new Buffer to fsys tree.
func (fsys *FS) Create(p string, m os.FileMode) (fs.File, error) {
	// TODO: split p in dir and file name. Note that as soon
	// as we're adding a Buffer, we need to create the necessary
	// intermediate directories if needed. The "file" component
	// of the path *must* be present.
	return nil, errors.New("create: not implemented")
}

func (fsys *FS) Remove(p string) error {
	// TODO: we need a strict policy for deciding what should be
	// deleted and what should not.
	return errors.New("remove: not implemented")
}
