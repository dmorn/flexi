// SPDX-FileCopyrightText: 2020 jecoz
//
// SPDX-License-Identifier: BSD-3-Clause

package file

import (
	"bytes"
	"io"
	"os"
	"sync"
	"time"
)

type Multi struct {
	name string

	sync.RWMutex
	buf     *LimitBuffer
	modTime time.Time
}

// Write will make Multi act as an io.MultiWriter on
// each Writer in m.writers.
func (m *Multi) Write(p []byte) (int, error) {
	m.Lock()
	defer m.Unlock()
	m.modTime = time.Now()
	return m.buf.Write(p)
}

func (m *Multi) Close() error {
	m.Lock()
	defer m.Unlock()
	return m.buf.Close()
}

func (m *Multi) Open() (io.ReadWriteCloser, error) {
	m.Lock()
	defer m.Unlock()
	return &LimitBuffer{
		buf: *bytes.NewBuffer(m.buf.Bytes()),
	}, nil
}

func (m *Multi) Stat() (os.FileInfo, error) {
	m.RLock()
	defer m.RUnlock()

	mtime := m.modTime
	if mtime.IsZero() {
		mtime = time.Now()
	}

	return Info{
		name:    m.name,
		size:    m.buf.Size(),
		mode:    0444,
		modTime: mtime,
		isDir:   false,
	}, nil
}

func NewMulti(name string) *Multi {
	return &Multi{
		name:    name,
		buf:     &LimitBuffer{},
		modTime: time.Now(),
	}
}
