// SPDX-FileCopyrightText: 2020 jecoz
//
// SPDX-License-Identifier: BSD-3-Clause

package file

import (
	"errors"
	"fmt"
	"io"
	"os"
	"sync"
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
	Name string

	sync.Mutex
	ModTime time.Time
	ReadAlt func([]byte) (int, error)
}

func (h *HackableRead) Close() error {
	h.Lock()
	defer h.Unlock()
	h.ReadAlt = func([]byte) (int, error) {
		return 0, io.EOF
	}
	h.ModTime = time.Now()
	return nil
}
func (h *HackableRead) Open() (io.ReadWriteCloser, error) {
	h.Lock()
	defer h.Unlock()
	var lastErr error
	return &HackableRWC{
		ReadAlt: func(p []byte) (int, error) {
			h.Lock()
			defer h.Unlock()
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

			h.ModTime = time.Now()
			n, err := h.ReadAlt(p)
			lastErr = err
			return n, err
		},
	}, nil
}
func (h *HackableRead) Stat() (os.FileInfo, error) {
	h.Lock()
	defer h.Unlock()
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
