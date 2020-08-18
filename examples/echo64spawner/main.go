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

	spawner := &fargate.Fargate{Backup: false}
	rp, err := spawner.Spawn(ctx, os.Stdin, 0)
	if err != nil {
		exitf("spawn task: %v", err)
	}
	if _, err := io.Copy(os.Stdout, rp.SpawnedReader()); err != nil {
		exitf("output copy: %v", err)
	}
}
