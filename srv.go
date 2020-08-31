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

	"aqwari.net/net/styx"

)

type Srv struct {
	S    Spawner
	pool *idPool
}

func (s *Srv) Serve(ln net.Listern) error {
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

func Serve(ln net.Listener, s Spawner) error {
	srv := &Srv{S: s, pool: new(idPool)}

	// Now retrieve remote processes that are still
	// running and try mounting them back.

	remotes, err := s.Ls()
	if err != nil {
		return err
	}
	restored := 0
	for i, v := range remotes {
		if err = srv.Restore(v); err != nil {
			log.Printf("error * restore failed (%d): %v", i, err)
			continue
		}
		restored++
	}
	log.Printf("*** %d remotes restored", restored)

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

	log.Printf("*** listening on %v", ln.Addr())
	return srv.Serve()
}

