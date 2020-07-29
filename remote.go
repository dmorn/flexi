package flexi

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/jecoz/flexi/file"
	"github.com/jecoz/flexi/fs"
)

type Remote struct {
	*file.Dir
	Spawner

	Done func(*Remote)
	mtpt string
}

func (r *Remote) mount(ctx context.Context, path string, stdio *Stdio) {
	h := &JsonHelper{}
	herr := func(err error) {
		h.Err(stdio.Err, err)
	}
	status := func(format string, args ...interface{}) {
		fmt.Fprintf(stdio.Status, format+"\n", args...)
	}

	task, err := DecodeTask(stdio.In)
	if err != nil {
		herr(err)
		return
	}
	status("task %v decoded", task.ID)

	rp, err := r.Spawn(ctx, task)
	if err != nil {
		herr(err)
		return
	}
	status("remote process %v spawned @ %v", rp.Name, rp.Addr)

	// From now on we also need to remove the spawned
	// process in case of error to avoid resource leaks.
	oldherr := herr
	herr = func(err error) {
		r.Kill(ctx, rp)
		oldherr(err)
		return
	}

	// TODO: OS dependent!
	cmd := exec.CommandContext(ctx, "9", "mount", rp.Addr, path)
	if err := cmd.Run(); err != nil {
		herr(err)
		return
	}
	status("remote process %v mounted @ %v", rp.Name, path)

	oldherr = herr
	herr = func(err error) {
		exec.CommandContext(ctx, "umount", path).Run()
		os.RemoveAll(path)
		oldherr(err)
		return
	}

	spawned, err := os.Create(filepath.Join(path, "spawned"))
	if err != nil {
		herr(err)
		return
	}
	defer spawned.Close()

	if err = EncodeRemoteProcess(spawned, rp); err != nil {
		herr(err)
		return
	}
	status("remote process info encoded & saved")
}

func NewRemote(mtpt, name string, s Spawner) (*Remote, error) {
	// First check that the file is not present already.
	// In that case, it means this remote should've been
	// restored instead, or might be. Anyway it **might**
	// not be treated as an error in the future.
	path := filepath.Join(mtpt, name)
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		return nil, fmt.Errorf("remote exists already at %v", path)
	}

	r := &Remote{mtpt: mtpt, Spawner: s}
	errfile := file.NewMulti("err")
	statusfile := file.NewMulti("status")

	spawn := file.NewPlumber("spawn", func(p *file.Plumber) bool {
		go func() {
			defer errfile.Close()
			defer statusfile.Close()

			r.mount(context.Background(), path, &Stdio{
				In:     p,
				Err:    errfile,
				Status: statusfile,
			})
		}()
		return true
	})
	static := []fs.File{spawn, errfile, statusfile}
	r.Dir = file.NewDirLS(name, func() []fs.File {
		return append(static, file.DiskLS(path)()...)
	})
	return r, nil
}
