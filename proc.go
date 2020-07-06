package flexi

import "io"

type Stdio struct {
	// In contains the input bytes. For 9p processes,
	// its the data written to the ctl file, which triggers
	// the execution of the process.
	In io.Reader
	// Write here execution errors.
	Err func() io.WriteCloser
	// Write here final output.
	Retv func() io.WriteCloser
}

// Processor describes an entity that is capable of executing
// a task reading and writing from Stdio.
type Processor interface {
	Run(*Stdio)
}

// ProcessorFunc is an helper Processor implementation that
// allows to make Processor even ordinary functions.
type ProcessorFunc func(*Stdio)

func (f ProcessorFunc) Run(i *Stdio) { f(i) }
