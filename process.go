package flexi

import (
	"io"
)

type Process interface {
	Ctl() io.Reader
	Status() io.Writer
	Retv() io.Writer
	Err() io.Writer
}

type Processor func(Process)
