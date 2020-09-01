package synth

import (
	"fmt"
	"io"

	"github.com/jecoz/flexi/fs"
)

type hackedFile struct {
	err error
	fs.File
	altRead  func([]byte) (int, error)
	altClose func() error
}

func (f *hackedFile) Read(p []byte) (int, error) {
	if a := f.altRead; a != nil {
		n, err := a(p)
		f.err = err
		return n, err
	}
	return f.File.Read(p)
}

func (f *hackedFile) Close() error {
	if f.altClose != nil {
		return f.altClose()
	}
	return f.File.Close()
}

func (f *hackedFile) Write(p []byte) (int, error) {
	if w, ok := f.File.(io.Writer); ok {
		return w.Write(p)
	}
	return 0, fmt.Errorf("not supported")
}

func HackClose(o Opener, alt func() error) Opener {
	return OpenerFunc(func() (fs.File, error) {
		f, err := o.Open()
		if err != nil {
			return f, err
		}
		return &hackedFile{
			File:     f,
			altClose: alt,
		}, nil
	})
}

func HackRead(o Opener, alt func([]byte) (int, error)) Opener {
	return OpenerFunc(func() (fs.File, error) {
		f, err := o.Open()
		if err != nil {
			return f, err
		}
		return &hackedFile{
			File:    f,
			altRead: alt,
		}, nil
	})
}
