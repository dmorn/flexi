package process

import (
	"net"
	"fmt"
	"strings"

	"github.com/Harvey-OS/ninep/protocol"
)

type Proxy struct {
	Addr net.Addr
	client *protocol.Client
}

func (p *Proxy) Dial() (net.Conn, error) {
	return net.Dial(p.Addr.Network(), p.Addr.String())
}

func (p *Proxy) IsConnected() bool { return p.client != nil }

// Proxy dials a 9P connection to p.Addr and exchanges the "version"
// transaction. If everything goes well, p.client will be filled with
// a ready-to-use 9P client interfacing the remote process.
func (p *Proxy) Connect() error {
	conn, err := p.Dial()
	if err != nil {
		return err
	}
	client, err := protocol.NewClient(func(c *protocol.Client) error {
		c.ToNet, c.FromNet = conn, conn
		return nil
	})
	if err != nil {
		return err
	}
	undo := true
	defer func() {
		if !undo { return }
		conn.Close()
		client.Dead = true
	}()

	msize := protocol.MaxSize(protocol.MSIZE)
	rmsize, rversion, err := client.CallTversion(msize, "9P2000")
	if err != nil {
		return err
	}
	if rmsize > msize {
		return fmt.Errorf("max message size mismatch: want <= %d, got %d", msize, rmsize)
	}
	if !strings.HasPrefix(rversion, "9P") {
		return fmt.Errorf("unsupported server version (%v)", rversion)
	}

	p.client, undo = client, false
	return nil
}
