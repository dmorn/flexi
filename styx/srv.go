// SPDX-FileCopyrightText: 2020 jecoz
//
// SPDX-License-Identifier: BSD-3-Clause

package styx

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"aqwari.net/net/styx"
	"github.com/jecoz/flexi"
)

type Srv struct {
	Mtpt string
	Ln   net.Listener
	S    flexi.Spawner

	index int
	fs    *fs
}

func (s *Srv) Serve() error {
	echo := styx.HandlerFunc(func(s *styx.Session) {
		for s.Next() {
			log.Printf("%q %T %s", s.User, s.Request(), s.Request().Path())
		}
	})
	srv := &styx.Server{
		Handler:  styx.Stack(echo, s),
		ErrorLog: log.New(os.Stderr, "", log.LstdFlags),
	}
	return srv.Serve(s.Ln)
}

func (srv *Srv) Serve9P(s *styx.Session) {
	for s.Next() {
		srv.fs.serveRequest(s.Request())
	}
}

func (srv *Srv) spawn(ctx context.Context, rc io.ReadCloser) error {
	defer rc.Close()

	var task flexi.Task
	if err := json.NewDecoder(rc).Decode(&task); err != nil {
		return fmt.Errorf("spawn process: %w", err)
	}
	rp, err := srv.S.Spawn(ctx, task)
	if err != nil {
		return err
	}

	fmt.Println("SPAWNED")
	if err := json.NewEncoder(os.Stdout).Encode(rp); err != nil {
		return err
	}
	name := fmt.Sprintf("%d", srv.index)
	mtpt := filepath.Join(srv.Mtpt, name)
	fmt.Printf("name: %v, mptp: %v\n", name, mtpt)
	cmd := exec.Command("9", "mount", rp.Addr, mtpt)
	if err := cmd.Run(); err != nil {
		return err
	}
	srv.index++

	fmt.Printf("MOUNTED @ %v\n", mtpt)
	return nil
}

func (srv *Srv) SpawnProcess(rc io.ReadCloser) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*5)
	defer cancel()
	if err := srv.spawn(ctx, rc); err != nil {
		log.Printf("ERROR * %v", err)
		return err
	}
	return nil
}

// ServeFlexi creates a styx process serving 9p over ln. To gracefully shutdown
// the server, close the listener.
func ServeFlexi(ln net.Listener, mntpt string, s flexi.Spawner) error {
	srv := &Srv{Mtpt: mntpt, Ln: ln, S: s}
	ctl := NewInputBuffer("ctl", srv.SpawnProcess)
	root := &Dir{
		Name:    "",
		Perm:    os.ModePerm,
		ModTime: time.Now(),
		Ls: func() []File {
			union := append([]File{ctl}, DiskLs(mntpt)()...)
			return union
		},
	}
	srv.fs = &fs{Root: root}

	return srv.Serve()
}
