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

// Multi is a fs.File implementation that broadcasts
// whatever is written to it to all its Open readers.
// Before Close is called, each Open call returns a stream
// reader that never returs io.EOF buf rather blocks.
// When Close is called, all readers are unlocked and will
// eventually reach the end of the stream. After this point,
// Open simply returns a buffer containing its contents.
// Initialize it with NewMulti.
type Multi struct {
	name string

	sync.RWMutex
	buf  *LimitBuffer
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

// Close closes the output streams. Subsequent Read calls on the opened
// streams will return io.EOF. Close stops if it is not capable of closing
// a stream, but removes the writers that have been successfully closed.
// Keeping on calling Close till nil is returned will eventually close
// every stream.
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
