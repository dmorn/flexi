// SPDX-FileCopyrightText: 2020 jecoz
//
// SPDX-License-Identifier: BSD-3-Clause

package file

import (
	"bytes"
	"errors"
)

type LimitBuffer struct {
	Max int64
	buf bytes.Buffer
}

func (b *LimitBuffer) Close() error               { return nil }
func (b *LimitBuffer) Read(p []byte) (int, error) { return b.buf.Read(p) }

// Write writes at most f.Max bytes inside the f.buf.
// If f.Max is not provided, no boundaries are explicitly set.
func (b *LimitBuffer) Write(p []byte) (int, error) {
	if b.Max > 0 && int64(len(p))+b.Size() > b.Max {
		n := b.Max - b.Size()
		written, err := b.buf.Write(p[:n])
		if err != nil {
			return written, err
		}
		return written, errors.New("buffer is full")
	}
	return b.buf.Write(p)
}

func (b *LimitBuffer) Bytes() []byte { return b.buf.Bytes() }
func (b *LimitBuffer) Size() int64   { return int64(b.buf.Len()) }
