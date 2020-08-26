// SPDX-FileCopyrightText: 2020 jecoz
//
// SPDX-License-Identifier: BSD-3-Clause

package process

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strconv"

	"aqwari.net/net/styx"
	"github.com/jecoz/flexi/fs"
	"github.com/jecoz/flexi/fs/styxfs"
	"github.com/jecoz/flexi/fs/synthfs"
)

type Stdio struct {
	// In contains the input bytes. For 9p processes,
	// its the data written to the in file, which triggers
	// the execution of the process.
	In io.Reader
	// Write here execution errors.
	Err io.WriteCloser
	// Write here final output.
	Retv io.WriteCloser
	// Write here status updates.
	State io.WriteCloser
}

// Runner describes an entity that is capable of executing
// a task reading and writing from Stdio.
type Runner interface {
	// Stdio Err and Retv buffers should not be used
	// after Run returns.
	Run(*Stdio)
}

// RunnerFunc is an helper Runner implementation that
// allows to make Runner even ordinary functions.
type RunnerFunc func(*Stdio)

func (f RunnerFunc) Run(i *Stdio) { f(i) }

type Srv struct {
	*styxfs.FS
}

func (s *Srv) Serve(ln net.Listener) error {
	srv := &styx.Server{
		// Remember that it is possible to stack handlers
		// using the styx.Stack helper.
		Handler:  s,
		TraceLog: log.New(os.Stderr, "", log.Ltime),
	}
	return srv.Serve(ln)
}

func NewSrv(fsys fs.RWFS) *Srv { return &Srv{styxfs.New(fsys)} }

func addBuffer(fsys *synthfs.FS, n string, m os.FileMode, in ...string) *synthfs.Buffer {
	b := &synthfs.Buffer{
		Name: n,
		Mode: m,
	}
	if err := fsys.AddOpener(b, n, in...); err != nil {
		panic(err)
	}
	return b
}

func Serve(ln net.Listener, r Runner) error {
	fsys := new(synthfs.FS)
	addBuffer(fsys, "hello", os.ModePerm)

	log.Printf("*** listening on %v", ln.Addr())
	return NewSrv(fsys).Serve(ln)
}

// Use NewHelper to create a working instance of Helper.
type Helper struct {
	tot float64
	i   *Stdio
	pw  *csv.Writer
}

func (h *Helper) relayErr(err error) {
	// TODO: panic-ing here is just a tmp solution.
	// I would rather prefer to contact a human.
	// Log? Email? Slack?
	panic(err)
}

func (h *Helper) Progress(step int, format string, args ...interface{}) {
	if err := h.pw.Write([]string{
		strconv.FormatFloat(float64(step)/h.tot, 'f', -1, 64),
		fmt.Sprintf(format, args...),
	}); err != nil {
		h.relayErr(err)
		return
	}
	h.pw.Flush()
	if err := h.pw.Error(); err != nil {
		h.relayErr(err)
	}
}

func (h *Helper) Err(err error) {
	// TODO: if we write multiple times to h.Err we'll produce
	// and invalid json payload. We should be able to truncate
	// the file instead.
	if werr := json.NewEncoder(h.i.Err).Encode(&struct {
		Error string `json:"error"`
	}{
		Error: err.Error(),
	}); werr != nil {
		h.relayErr(fmt.Errorf("%v: %w", werr, err))
	}
}

func (h *Helper) Errf(format string, args ...interface{}) {
	h.Err(fmt.Errorf(format, args...))
}

func (h *Helper) Retv(v interface{}) {
	if err := json.NewEncoder(h.i.Retv).Encode(v); err != nil {
		// Try telling the user about the error!
		h.Err(err)
	}
}

func (h *Helper) JSONDecodeInput(v interface{}) error {
	return json.NewDecoder(h.i.In).Decode(v)
}

// Done writes the final "Done" message, indicating that the process
// finished doing its task and will not post any more status update.
func (h *Helper) Done() { h.Progress(int(h.tot), "done!") }

func NewHelper(i *Stdio, tot int) *Helper {
	return &Helper{
		tot: float64(tot),
		i:   i,
		pw:  csv.NewWriter(i.State),
	}
}
