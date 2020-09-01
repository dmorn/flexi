// SPDX-FileCopyrightText: 2020 jecoz
//
// SPDX-License-Identifier: BSD-3-Clause

package flexi

import (
	"io"
	"log"
	"net"
	"os"

	"aqwari.net/net/styx"
	"github.com/jecoz/flexi/styx/styxfs"
	"github.com/jecoz/flexi/synth"
)

type Srv struct {
	*styxfs.FS
	S    Spawner
	pool *idPool
}

func (s *Srv) Serve(ln net.Listener) error {
	srv := &styx.Server{
		Handler:  s.FS,
		TraceLog: log.New(os.Stderr, "", log.Ltime),
	}
	return srv.Serve(ln)
}

// func (s *Srv) addRemote(id int, f func(string, int) (*Remote, error)) (*Remote, error) {
// 	if id < 0 {
// 		id = s.pool.Get()
// 	} else {
// 		// Notify the pool that we have this id
// 		// already and no other Get() call should
// 		// return id till we Put it back to the pool.
// 		if err := s.pool.Have(id); err != nil {
// 			return nil, fmt.Errorf("add remote: invalid id requested: %w", err)
// 		}
// 	}
// 	r, err := f(strconv.Itoa(id), id)
// 	if err != nil {
// 		s.pool.Put(id)
// 		return nil, err
// 	}
// 	r.Done = func() {
// 		s.pool.Put(id)
// 	}
// 	return r, nil
// }
//
// func (s *Srv) NewRemote() (*Remote, error) {
// 	return s.addRemote(-1, func(name string, id int) (*Remote, error) {
// 		return NewRemote(s.Mtpt, name, s.S, id)
// 	})
// }
//
// func (s *Srv) RestoreRemote(rp *RemoteProcess) (*Remote, error) {
// 	return s.addRemote(rp.ID, func(name string, id int) (*Remote, error) {
// 		return RestoreRemote(s.Mtpt, name, s.S, rp)
// 	})
// }

func Serve(ln net.Listener, s Spawner) error {
	srv := &Srv{S: s, pool: new(idPool)}
	// Now retrieve remote processes that are still
	// running and try mounting them back.

	remotes, err := s.Ls()
	if err != nil {
		return err
	}
	restored := 0
	for _, _ = range remotes {
		//if err = srv.Restore(v); err != nil {
		//	log.Printf("error * restore failed (%d): %v", i, err)
		//	continue
		//}
		//restored++
	}
	log.Printf("*** %d remotes restored", restored)

	fsys := new(synth.FS)

	cloneb := &synth.Buffer{Name: "clone", Mode: 0440}
	clone := synth.HackRead(cloneb, func(p []byte) (int, error) {
		// TODO: implement
		return copy(p, []byte("mother")), io.EOF
		// Users read the clone file to obtain

		// a new remote process.
		// remote, err := srv.NewRemote()
		// if err != nil {
		// 	return 0, err
		// }

		// s := []byte(remote.Name + "\n")
		// if len(s) > len(p) {
		// 	remote.Done()
		// 	return 0, io.ErrShortBuffer
		// }

		// srv.FS.Create("", remote)
		// return copy(p, s), io.EOF

	})

	if err := fsys.InsertOpener(clone); err != nil {
		return err
	}
	log.Printf("*** listening on %v", ln.Addr())
	srv.FS = styxfs.New(fsys)
	return srv.Serve(ln)
}
