// SPDX-FileCopyrightText: 2020 jecoz
//
// SPDX-License-Identifier: BSD-3-Clause

package flexi

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"sync"

	"github.com/jecoz/flexi/file"
	"github.com/jecoz/flexi/file/memfs"
	"github.com/jecoz/flexi/fs"
	"github.com/jecoz/flexi/styx"
)

type Srv struct {
	Mtpt string
	Ln   net.Listener
	S    Spawner
	FS   fs.FS

	pool *intPool
}

func (s *Srv) Serve() error {
	return styx.Serve(s.Ln, s.FS)
}

func (s *Srv) addRemote(f func(string) (*Remote, error)) (*Remote, error) {
	index := s.pool.Get()
	r, err := f(strconv.Itoa(int(index)))
	if err != nil {
		s.pool.Put(index)
		return nil, err
	}
	r.Done = func() {
		s.pool.Put(index)
	}
	return r, nil
}

func (s *Srv) NewRemote() (*Remote, error) {
	return s.addRemote(func(name string) (*Remote, error) {
		return NewRemote(s.Mtpt, name, s.S)
	})
}

func (s *Srv) RestoreRemote(rp *RemoteProcess) (*Remote, error) {
	return s.addRemote(func(name string) (*Remote, error) {
		return RestoreRemote(s.Mtpt, name, s.S, rp)
	})
}

func (s *Srv) cleanupMtpt() error {
	for i, v := range file.DiskLS(s.Mtpt)() {
		info, err := v.Stat()
		if err != nil {
			return fmt.Errorf("clean-up mtpt (%d): %v", i, err)
		}
		path := filepath.Join(s.Mtpt, info.Name())

		// We do not care if the operation is not successfull.
		// It might also be that there is nothing to umount.
		umount(path)
		if err = os.RemoveAll(path); err != nil {
			return fmt.Errorf("clean-up mtpt (%d): %v", i, err)
		}
	}
	return nil
}

func ServeFlexi(ln net.Listener, mtpt string, s Spawner) error {
	srv := &Srv{Mtpt: mtpt, Ln: ln, S: s, pool: newIntPool()}

	// Start from a clean state, otherwise we could encounter
	// issues later on.
	if err := srv.cleanupMtpt(); err != nil {
		return err
	}

	// Now retrieve remote processes that are still
	// running and try mounting them back.

	oldremotes, err := s.LS()
	if err != nil {
		return err
	}
	remotes := make([]*Remote, 0, len(oldremotes))
	for i, v := range oldremotes {
		restored, err := srv.RestoreRemote(v)
		if err != nil {
			log.Printf("error * restore failed (%d): %v", i, err)
			continue
		}
		remotes = append(remotes, restored)
	}
	log.Printf("*** %d remotes restored from %v", len(remotes), mtpt)

	clone := file.WithRead("clone", func(p []byte) (int, error) {
		// Users read the clone file to obtain
		// a new remote process.
		remote, err := srv.NewRemote()
		if err != nil {
			return 0, err
		}

		s := []byte(remote.Name + "\n")
		if len(s) > len(p) {
			remote.Done()
			return 0, io.ErrShortBuffer
		}

		srv.FS.Create("", remote)
		return copy(p, s), io.EOF
	})
	files := append(make([]fs.File, 0, len(remotes)+1), clone)
	for _, v := range remotes {
		files = append(files, v)
	}
	root := file.NewDirFiles("", files...)
	srv.FS = memfs.New(root)

	return srv.Serve()
}

type intPool struct {
	n    int64
	pool sync.Pool
}

func newIntPool() (p *intPool) {
	p = new(intPool)
	p.pool = sync.Pool{
		New: func() interface{} {
			defer func() { p.n++ }()
			return p.n
		},
	}
	return
}

func (p *intPool) Get() int64  { return p.pool.Get().(int64) }
func (p *intPool) Put(i int64) { p.pool.Put(i) }
