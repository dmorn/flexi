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

func NewRemote(mtpt, name string, s Spawner) (*Remote, error) {
	// First check that the file is not present already.
	// In that case, it means this remote should've been
	// restored instead, or might be. Anyway it **might**
	// not be treated as an error in the future.
	path := filepath.Join(mtpt, name)
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		return nil, fmt.Errorf("remote exists already at %v", path)
	}

	errfile := file.NewMulti("err")
	h := &JsonHelper{}
	herr := func(err error) bool {
		h.Err(errfile, err)
		return false
	}

	spawn := file.NewPlumber("spawn", func(p *file.Plumber) bool {
		defer errfile.Close()

		task, err := DecodeTask(p)
		if err != nil {
			return herr(err)
		}
		ctx := context.Background()
		rp, err := s.Spawn(ctx, *task)
		if err != nil {
			return herr(err)
		}

		// From now on we also need to remove the spawned
		// process in case of error to avoid resource leaks.
		herr = func(err error) bool {
			s.Kill(ctx, rp)
			return herr(err)
		}

		// TODO: OS dependent!
		cmd := exec.CommandContext(ctx, "9", "mount", rp.Addr, path)
		if err := cmd.Run(); err != nil {
			return herr(err)
		}
		herr = func(err error) bool {
			exec.CommandContext(ctx, "umount", path).Run()
			os.RemoveAll(path)
			return herr(err)
		}

		spawned, err := os.Create(filepath.Join(path, "spawned"))
		if err != nil {
			return herr(err)
		}
		defer spawned.Close()

		if err = EncodeRemoteProcess(spawned, rp); err != nil {
			return herr(err)
		}
		return true
	})

	return &Remote{
		mtpt:    mtpt,
		Spawner: s,
		Dir: file.NewDirLS(name, func() []fs.File {
			return append([]fs.File{spawn, errfile}, file.DiskLS(path)()...)
		}),
	}, nil
}
