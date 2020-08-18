// SPDX-FileCopyrightText: 2020 jecoz
//
// SPDX-License-Identifier: BSD-3-Clause

package flexi

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"strconv"

	"github.com/jecoz/flexi/file"
	"github.com/jecoz/flexi/file/memfs"
	"github.com/jecoz/flexi/fs"
	"github.com/jecoz/flexi/styx"
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

// Processor describes an entity that is capable of executing
// a task reading and writing from Stdio.
type Processor interface {
	// Stdio Err and Retv buffers should not be used
	// after Run returns.
	Run(*Stdio)
}

// ProcessorFunc is an helper Processor implementation that
// allows to make Processor even ordinary functions.
type ProcessorFunc func(*Stdio)

func (f ProcessorFunc) Run(i *Stdio) { f(i) }

type Process struct {
	FS     fs.FS
	Ln     net.Listener
	Runner Processor
}

func (p *Process) Serve() error {
	return styx.Serve(p.Ln, p.FS)
}

func ServeProcess(ln net.Listener, r Processor) error {
	err := file.NewMulti("err")
	retv := file.NewMulti("retv")
	state := file.NewMulti("state")
	in := file.NewPlumber("in", func(p *file.Plumber) bool {
		buf := new(bytes.Buffer)
		if _, err := io.Copy(buf, p); err != nil {
			return false
		}

		go func() {
			stdio := &Stdio{
				In:    buf,
				Err:   err,
				Retv:  retv,
				State: state,
			}
			r.Run(stdio)
			err.Close()
			retv.Close()
		}()
		return true
	})
	root := file.NewDirFiles(
		"",
		in,
		err,
		retv,
		state,
	)
	p := Process{
		FS:     memfs.New(root),
		Ln:     ln,
		Runner: r,
	}
	log.Printf("*** listening on %v", ln.Addr())
	return p.Serve()
}

// Use NewProcessHelper to create a working instance
// of ProcessHelper.
type ProcessHelper struct {
	tot float64
	i   *Stdio
	pw  *csv.Writer
}

func (h *ProcessHelper) relayErr(err error) {
	// TODO: panic-ing here is just a tmp solution.
	// I would rather prefer to contact a human.
	// Log? Email? Slack?
	panic(err)
}

func (h *ProcessHelper) Progress(step int, format string, args ...interface{}) {
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

func (h *ProcessHelper) Err(err error) {
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

func (h *ProcessHelper) Errf(format string, args ...interface{}) {
	h.Err(fmt.Errorf(format, args...))
}

func (h *ProcessHelper) Retv(v interface{}) {
	if err := json.NewEncoder(h.i.Retv).Encode(v); err != nil {
		// Try telling the user about the error!
		h.Err(err)
	}
}

func (h *ProcessHelper) JSONDecodeInput(v interface{}) error {
	return json.NewDecoder(h.i.In).Decode(v)
}

// Done writes the final "Done" message, indicating that the process
// finished doing its task and will not post any more status update.
func (h *ProcessHelper) Done() { h.Progress(int(h.tot), "done!") }

func NewProcessHelper(i *Stdio, tot int) *ProcessHelper {
	return &ProcessHelper{
		tot: float64(tot),
		i:   i,
		pw:  csv.NewWriter(i.State),
	}
}
