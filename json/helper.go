// SPDX-FileCopyrightText: 2020 jecoz
//
// SPDX-License-Identifier: BSD-3-Clause

package json

import (
	"encoding/json"
	"io"

	"github.com/jecoz/flexi"
)

type ProcessHelper struct {
	*flexi.Stdio
}

func encode(w io.WriteCloser, v interface{}) error {
	defer w.Close()
	return json.NewEncoder(w).Encode(v)
}

func (h *ProcessHelper) Retv(v interface{}) error {
	return encode(h.Stdio.Retv(), v)
}

func (h *ProcessHelper) Err(err error) error {
	return encode(h.Stdio.Err(), &struct {
		Error string `json:"error"`
	}{
		Error: err.Error(),
	})
}