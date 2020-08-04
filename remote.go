// SPDX-FileCopyrightText: 2020 jecoz
//
// SPDX-License-Identifier: BSD-3-Clause

package flexi

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/jecoz/flexi/file"
	"github.com/jecoz/flexi/fs"
)

type Remote struct {
	*file.Dir
	Spawner
	Name string
	Done func()

	mtpt    string
	spawned io.Reader
}

func (r *Remote) Close() error {
	if r.spawned != nil {
		mtpt := filepath.Join(r.mtpt, r.Name)
		if err := Umount(mtpt); err != nil {
			return fmt.Errorf("unable to umount %v: %w", mtpt, err)
		}
		if err := r.Kill(context.Background(), r.spawned); err != nil {
			return err
		}
	}
	r.Dir = file.NewDirFiles("")
	if r.Done != nil {
		r.Done()
	}
	return nil
}

func Mount(addr, mtpt string) error {
	return mount(addr, mtpt)
}

func Umount(path string) error {
	if err := umount(path); err != nil {
		return err
	}
	return os.RemoveAll(path)
}

func (r *Remote) mirrorRemoteProcess(ctx context.Context, path string, i *Stdio) {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()

	// Prepare output encoding helpers. If this is the behaviour
	// of every flexi process, we could add one more helper layer.

	h := NewProcessHelper(i, 7)
	defer h.Done()
	herr := func(err error) {
		h.Err(err)
	}

	h.Progress(1, "starting %v mount process", path)
	h.Progress(2, "spawning remote process")
	rp, err := r.Spawn(ctx, i.In)
	if err != nil {
		herr(err)
		return
	}
	h.Progress(3, "remote process spawned @ %v", rp.Addr)

	// From now on we also need to remove the spawned
	// process in case of error to avoid resource leaks.
	oldherr := herr
	herr = func(err error) {
		r.Kill(ctx, rp.Spawned)
		oldherr(err)
	}

	if err := Mount(rp.Addr, path); err != nil {
		herr(err)
		return
	}
	h.Progress(4, "remote process mounted @ %v", path)

	oldherr = herr
	herr = func(err error) {
		exec.CommandContext(ctx, "umount", path).Run()
		os.RemoveAll(path)
		oldherr(err)
	}

	h.Progress(5, "storing spawn information at %v", path)

	// TODO: try creating a version of this function that can
	// detect when it is not possible to create the file in the
	// remote namespace w/o leaking goroutines nor locking.
	spawned, err := os.Create(filepath.Join(path, "spawned"))
	if err != nil {
		herr(err)
		return
	}
	defer spawned.Close()

	// If we read straight from rp.Spawned we'll consume its contents.
	// This way we read inside b (which we use internally) and
	// inside the spawned file.

	var b bytes.Buffer
	tee := io.TeeReader(rp.Spawned, &b)
	if _, err := io.Copy(spawned, tee); err != nil {
		herr(err)
		return
	}
	r.spawned = &b
	h.Progress(6, "remote process info encoded & saved")
}

func RestoreRemote(mtpt string, name string, s Spawner, rp *RemoteProcess) (*Remote, error) {
	// In contrast with NewRemote, we're not killing anything
	// here even though we could.
	path := filepath.Join(mtpt, name)
	if err := Mount(rp.Addr, path); err != nil {
		return nil, err
	}

	// We assume we're restoring a spawned remote. If
	// that is the case, there is no need for creating
	// the err, state and spawn files, as they belong
	// to the past. If this wasn't a spawned remote,
	// users should just delete this and create a new one.

	mirror := file.NewDirLS("mirror", file.DiskLS(path))
	return &Remote{
		mtpt:    mtpt,
		Spawner: s,
		Name:    name,
		Dir:     file.NewDirFiles(name, mirror),
		spawned: rp.Spawned,
	}, nil
}

func NewRemote(mtpt string, name string, s Spawner) (*Remote, error) {
	// First check that the file is not present already.
	// In that case, it means this remote should've been
	// restored instead, or might be. Anyway it **might**
	// not be treated as an error in the future.
	path := filepath.Join(mtpt, name)
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		return nil, fmt.Errorf("remote exists already at %v", path)
	}

	r := &Remote{mtpt: mtpt, Spawner: s, Name: name}
	errfile := file.NewMulti("err")
	statefile := file.NewMulti("state")

	spawn := file.NewPlumber("spawn", func(p *file.Plumber) bool {
		go func() {
			defer errfile.Close()
			defer statefile.Close()

			r.mirrorRemoteProcess(context.Background(), path, &Stdio{
				In:    p,
				Err:   errfile,
				State: statefile,
			})
		}()
		return true
	})
	static := []fs.File{spawn, errfile, statefile}
	mirror := file.NewDirLS("mirror", file.DiskLS(path))
	r.Dir = file.NewDirFiles(name, append(static, mirror)...)
	return r, nil
}
