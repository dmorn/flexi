package process

import (
	"fmt"

	"github.com/Harvey-OS/ninep/protocol"
)

type Srv struct {
}

func (s *Srv) Rversion(rmsize protocol.MaxSize, rversion string) (protocol.MaxSize, string, error) {
	msize := protocol.MaxSize(protocol.MSIZE)
	version := "9P2000"
	if rmsize > msize {
		return 0, "", fmt.Errorf("max message size mismatch: want <= %d, got %d", msize, rmsize)
	}
	if rversion != version {
		return 0, "", fmt.Errorf("version (%s) is not supported, try with %s", rversion, version)
	}
	return rmsize, version, nil
}

func (s *Srv) Rattach(protocol.FID, protocol.FID, string, string) (protocol.QID, error) {
	return protocol.QID{}, fmt.Errorf("not implemented")
}

func (s *Srv) Rwalk(protocol.FID, protocol.FID, []string) ([]protocol.QID, error) {
	return nil, fmt.Errorf("not implemented")
}

func (s *Srv) Ropen(protocol.FID, protocol.Mode) (protocol.QID, protocol.MaxSize, error) {
	return protocol.QID{}, protocol.MaxSize(0), fmt.Errorf("not implemented")
}

func (s *Srv) Rcreate(protocol.FID, string, protocol.Perm, protocol.Mode) (protocol.QID, protocol.MaxSize, error) {
	return protocol.QID{}, protocol.MaxSize(0), fmt.Errorf("not implemented")
}

func (s *Srv) Rstat(protocol.FID) ([]byte, error) {
	return nil, fmt.Errorf("not implemented")
}

func (s *Srv) Rwstat(protocol.FID, []byte) error {
	return fmt.Errorf("not implemented")
}

func (s *Srv) Rclunk(protocol.FID) error {
	return fmt.Errorf("not implemented")
}

func (s *Srv) Rremove(protocol.FID) error {
	return fmt.Errorf("not implemented")
}

func (s *Srv) Rread(protocol.FID, protocol.Offset, protocol.Count) ([]byte, error) {
	return nil, fmt.Errorf("not implemented")
}

func (s *Srv) Rwrite(protocol.FID, protocol.Offset, []byte) (protocol.Count, error) {
	return protocol.Count(0), fmt.Errorf("not implemented")
}

func (s *Srv) Rflush(otag protocol.Tag) error {
	return fmt.Errorf("not implemented")
}
