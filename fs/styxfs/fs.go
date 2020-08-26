package styxfs

import (
	"errors"

	"aqwari.net/net/styx"
	"github.com/jecoz/flexi/fs"
)

type FS struct {
	fs.RWFS
}

type TruncateFile interface {
	fs.File
	Truncate(int64) error
}

func Truncate(f fs.File, s int64) error {
	if tf, ok := f.(TruncateFile); ok {
		return tf.Truncate(s)
	}
	return errors.New("Truncate not available")
}

func (fys *FS) handleT(t styx.Request) {
	switch msg := t.(type) {
	case styx.Topen:
		msg.Ropen(fys.Open(t.Path()))
	case styx.Tstat:
		file, err := fys.Open(t.Path())
		if err != nil {
			msg.Rstat(nil, err)
			return
		}
		msg.Rstat(file.Stat())
	case styx.Twalk:
		file, err := fys.Open(t.Path())
		if err != nil {
			msg.Rwalk(nil, err)
			return
		}
		msg.Rwalk(file.Stat())
	case styx.Tcreate:
		msg.Rcreate(fys.Create(msg.Path(), msg.Mode))
	case styx.Tremove:
		msg.Rremove(fys.Remove(msg.Path()))
	case styx.Ttruncate:
		file, err := fys.Open(msg.Path())
		if err != nil {
			msg.Rtruncate(err)
			return
		}
		msg.Rtruncate(Truncate(file, msg.Size))
	case styx.Tutimes:
		// Each file can handle this information without
		// requiring the user telling when the file has
		// been modified.
		msg.Rutimes(nil)
	default:
		// Default responses will take
		// care of the remaining/new messages.
	}
}

func (fys *FS) Serve9P(s *styx.Session) {
	for s.Next() {
		fys.handleT(s.Request())
	}
}

func New(p fs.RWFS) *FS { return &FS{p} }
