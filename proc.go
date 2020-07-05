package flexi

import "io"

type Stdio struct {
	In   io.Reader
	Err  func() io.WriteCloser
	Retv func() io.WriteCloser
}

type Processor interface {
	Run(*Stdio)
}

type ProcessorFunc func(*Stdio)

func (f ProcessorFunc) Run(i *Stdio) { f(i) }
