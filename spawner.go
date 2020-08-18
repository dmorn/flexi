// SPDX-FileCopyrightText: 2020 jecoz
//
// SPDX-License-Identifier: BSD-3-Clause

package flexi

import (
	"bytes"
	"context"
	"io"
)

type RemoteProcess struct {
	ID   int    `json:"id"`
	Addr string `json:"addr"`
	Name string `json:"name"`

	// Spawned contains the payload that needs to be preserved
	// in order to undo the Spawn operation. flexi does not
	// know how to encode/decode it, so it just saves it
	// inside files (in the remote fs itself and in a persistent
	// storage)
	Spawned []byte `json:"spawned"`
}

func (rp *RemoteProcess) SpawnedReader() io.Reader {
	b := make([]byte, len(rp.Spawned))
	copy(b, rp.Spawned)
	return bytes.NewReader(b)
}

type Spawner interface {
	Spawn(context.Context, io.Reader, int) (*RemoteProcess, error)
	Kill(context.Context, io.Reader) error
	Ls() ([]*RemoteProcess, error)
}
