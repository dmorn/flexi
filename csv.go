// SPDX-FileCopyrightText: 2020 jecoz
//
// SPDX-License-Identifier: BSD-3-Clause

package flexi

import (
	"encoding/csv"
	"io"
	"strconv"
)

type CSVProgressHelper struct {
	// Expected total number of steps.
	Tot float64

	w *csv.Writer

	// TODO: we could add an error handler delegate
	// to cope with SilentEncode errors.
}

func (h *CSVProgressHelper) Encode(v ...string) error {
	if err := h.w.Write(v); err != nil {
		return err
	}
	h.w.Flush()
	return h.w.Error()
}

func (h *CSVProgressHelper) SilentEncode(v ...string) {
	if err := h.Encode(v...); err != nil {
		// TODO: panic-ing here is not a solution.
		// I would rather prefer to contact a human.
		// Log? Email? Slack?
		panic(err)
	}
}

func (h *CSVProgressHelper) Progress(step int, description string) {
	p := float64(step) / h.Tot
	h.SilentEncode(strconv.FormatFloat(p, 'f', -1, 64), description)
}

// Done writes the final "Done" message, indicating that the process
// finished doing its task and will not post any more status update.
func (h *CSVProgressHelper) Done() { h.Progress(int(h.Tot), "done!") }

func NewCSVProgressHelper(w io.Writer, tot int) *CSVProgressHelper {
	return &CSVProgressHelper{Tot: float64(tot), w: csv.NewWriter(w)}
}
