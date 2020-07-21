// SPDX-FileCopyrightText: 2020 jecoz
//
// SPDX-License-Identifier: BSD-3-Clause

package flexi

import (
	"io"
)

type RwHook func(io.ReadWriter) error

type opCloser struct {
	io.ReadWriter
	h RwHook
}

func (c *opCloser) Close() error { return c.h(c) }

// OpCloser returns an io.ReadWriteCloser that calls f when Close is called.
// Close will return what f returns.
func OpCloser(rw io.ReadWriter, f RwHook) io.ReadWriteCloser {
	return &opCloser{rw, f}
}
