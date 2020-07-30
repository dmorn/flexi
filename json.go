// SPDX-FileCopyrightText: 2020 jecoz
//
// SPDX-License-Identifier: BSD-3-Clause

package flexi

import (
	"encoding/json"
	"io"
)

type JsonHelper struct {
	// TODO: we could add an error handler delegate
	// to cope with SilentEncode errors.
}

func (h *JsonHelper) SilentEncode(w io.Writer, v interface{}) {
	if err := h.Encode(w, v); err != nil {
		// TODO: panic-ing here is not a solution.
		// I would rather prefer to get in touch with
		// an Human.
		// Log? Email? Slack?
		panic(err)
	}
}

func (h *JsonHelper) Encode(w io.Writer, v interface{}) error {
	return json.NewEncoder(w).Encode(v)
}

func (h *JsonHelper) Err(w io.Writer, err error) {
	h.SilentEncode(w, &struct {
		Error string `json:"error"`
	}{
		Error: err.Error(),
	})
}
