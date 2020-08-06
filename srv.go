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
	"sort"
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

	pool *idPool
}

func (s *Srv) Serve() error {
	return styx.Serve(s.Ln, s.FS)
}

func (s *Srv) addRemote(id int, f func(string, int) (*Remote, error)) (*Remote, error) {
	if id < 0 {
		id = s.pool.Get()
	} else {
		// Notify the pool that we have this id
		// already and no other Get() call should
		// return id till we Put it back to the pool.
		if err := s.pool.Have(id); err != nil {
			return nil, fmt.Errorf("add remote: invalid id requested: %w", err)
		}
	}
	r, err := f(strconv.Itoa(id), id)
	if err != nil {
		s.pool.Put(id)
		return nil, err
	}
	r.Done = func() {
		s.pool.Put(id)
	}
	return r, nil
}

func (s *Srv) NewRemote() (*Remote, error) {
	return s.addRemote(-1, func(name string, id int) (*Remote, error) {
		return NewRemote(s.Mtpt, name, s.S, id)
	})
}

func (s *Srv) RestoreRemote(rp *RemoteProcess) (*Remote, error) {
	return s.addRemote(rp.ID, func(name string, id int) (*Remote, error) {
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
	srv := &Srv{Mtpt: mtpt, Ln: ln, S: s, pool: new(idPool)}

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

type idPool struct {
	sync.Mutex
	free []int
	out  []int
}

// Get returns a unique integer. Subsequent Get calls will not
// return the same integer unless it is returned to the pool
// with Put.
func (p *idPool) Get() int {
	p.Lock()
	defer p.Unlock()
	if p.out == nil {
		p.out = []int{}
	}
	if len(p.free) == 0 {
		max := 0
		if len(p.out) > 0 {
			max = p.out[0]
			max++
		}
		p.free = []int{max}
	}

	id := p.free[0]
	p.free = p.free[1:]
	p.out = append(p.out, id)
	sort.Sort(sort.Reverse(sort.IntSlice(p.out)))
	return id
}

// Put returns i to the pool, meaning that subsequent Get
// calls might return it.
func (p *idPool) Put(i int) {
	p.Lock()
	defer p.Unlock()

	remove := -1
	for j, v := range p.out {
		if v == i {
			remove = j
			break
		}
	}
	if remove == -1 {
		// Caller asked to remove an indentifier
		// that was not in the out list.
		// Is it critical?
		return
	}
	p.out = append(p.out[:remove], p.out[remove+1:]...)

	if p.free == nil {
		p.free = []int{}
	}
	p.free = append(p.free, i)
	sort.Ints(p.free)
}

// Have works like Get without returning any integer.
// Use it to let p know you have i, and no other should.
// Returns an error if p knew already that i was out.
func (p *idPool) Have(i int) error {
	p.Lock()
	defer p.Unlock()
	if p.out == nil {
		p.out = []int{}
	}
	for _, v := range p.out {
		if i == v {
			return fmt.Errorf("%d is already tracked by the pool", i)
		}
	}
	p.out = append(p.out, i)
	sort.Sort(sort.Reverse(sort.IntSlice(p.out)))
	return nil
}
