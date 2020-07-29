// SPDX-FileCopyrightText: 2020 jecoz
//
// SPDX-License-Identifier: BSD-3-Clause

package file

import (
	"errors"
	"fmt"
	"io"
	"os"
	"time"
)

type HackableRWC struct {
	ReadAlt func([]byte) (int, error)
}

var (
	ErrNotAllowed   = errors.New("operation not allowed")
	ReadNotAllowed  = fmt.Errorf("read: %w", ErrNotAllowed)
	WriteNotAllowed = fmt.Errorf("read: %w", ErrNotAllowed)
)

func (h *HackableRWC) Read(p []byte) (int, error) {
	if h.ReadAlt == nil {
		return 0, ReadNotAllowed
	}
	return h.ReadAlt(p)
}

func (h *HackableRWC) Write(p []byte) (int, error) {
	// If needed, add another alterative implementation.
	return 0, WriteNotAllowed
}

func (h *HackableRWC) Close() error { return nil }

type HackableRead struct {
	Name    string
	ModTime time.Time
	ReadAlt func([]byte) (int, error)
}

func (h *HackableRead) Open() (io.ReadWriteCloser, error) {
	var lastErr error
	return &HackableRWC{
		ReadAlt: func(p []byte) (int, error) {
			switch {
			case errors.Is(lastErr, io.EOF):
				return 0, lastErr
			case errors.Is(lastErr, io.ErrShortBuffer):
				// Maybe this time the user provided a
				// buffer that is big enough.
			case lastErr != nil:
				return 0, lastErr
			default:
			}

			n, err := h.ReadAlt(p)
			lastErr = err
			return n, err
		},
	}, nil
}
func (h *HackableRead) Stat() (os.FileInfo, error) {
	return Info{
		name:    h.Name,
		size:    0,
		mode:    0444,
		modTime: h.ModTime,
		isDir:   false,
	}, nil
}

func WithRead(name string, r func([]byte) (int, error)) *HackableRead {
	return &HackableRead{Name: name, ModTime: time.Now(), ReadAlt: r}
}
