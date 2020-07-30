// SPDX-FileCopyrightText: 2020 jecoz
//
// SPDX-License-Identifier: BSD-3-Clause

package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/jecoz/flexi/fargate"
)

func exitf(format string, args ...interface{}) {
	fmt.Printf("error * "+format+"\n", args...)
	os.Exit(1)
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()
	if err := new(fargate.Fargate).Kill(ctx, os.Stdin); err != nil {
		exitf("kill remote process: %v", err)
	}
}
