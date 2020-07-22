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

func encode(w io.Writer, v interface{}) error {
	return json.NewEncoder(w).Encode(v)
}

// Retv encodes and writes in json format v inside the internal
// retv WriteCloser. Closes the destination after usage.
func (h *ProcessHelper) Retv(v interface{}) error {
	return encode(h.Stdio.Retv, v)
}

// Retv encodes and writes in json format err inside the internal
// err WriteCloser. Closes the destination after usage.
func (h *ProcessHelper) Err(err error) error {
	return encode(h.Stdio.Err, &struct {
		Error string `json:"error"`
	}{
		Error: err.Error(),
	})
}
