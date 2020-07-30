// SPDX-FileCopyrightText: 2020 jecoz
//
// SPDX-License-Identifier: BSD-3-Clause

package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/jecoz/flexi/fargate"
)

func exitf(format string, args ...interface{}) {
	fmt.Printf("error * "+format+"\n", args...)
	os.Exit(1)
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*3)
	defer cancel()

	_, p, err := new(fargate.Fargate).Spawn(ctx, os.Stdin)
	if err != nil {
		exitf("spawn task: %v", err)
	}
	if _, err := io.Copy(os.Stdout, p); err != nil {
		exitf("output copy: %v", err)
	}
}
