// SPDX-FileCopyrightText: 2020 jecoz
//
// SPDX-License-Identifier: BSD-3-Clause

package flexi

import (
	"context"
	"io"
)

type RemoteProcess struct {
	Addr string
	Name string

	// Spawned contains the payload that needs to be preserved
	// in order to undo the Spawn operation. flexi does not
	// know how to encode/decode it, so it just saves it
	// inside files (in the remote fs itself and in a persistent
	// storage)
	Spawned io.Reader
}

type Spawner interface {
	Spawn(context.Context, io.Reader) (*RemoteProcess, error)
	Kill(context.Context, io.Reader) error
	LS() ([]*RemoteProcess, error)
}
