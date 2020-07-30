// SPDX-FileCopyrightText: 2020 jecoz
//
// SPDX-License-Identifier: BSD-3-Clause

package flexi

import (
	"context"
	"io"
)

type RemoteProcess struct {
	Addr string `json:"addr"`
	Name string `json:"name"`
}

type Spawner interface {
	Spawn(context.Context, io.Reader) (*RemoteProcess, io.Reader, error)
	Kill(context.Context, io.Reader) error
}
