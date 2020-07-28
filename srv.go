package flexi

import (
	"errors"
	"net"

	"github.com/jecoz/flexi/fs"
	"github.com/jecoz/flexi/styx"
)

type Srv struct {
	Ln net.Listener
	Spawner
	FS fs.FS
}

func (s *Srv) Serve() error {
	return styx.Serve(s.Ln, s.FS)
}

func ServeFlexi(ln net.Listener, mptp string, s Spawner) error {
	return errors.New("not implemented yet")
}
