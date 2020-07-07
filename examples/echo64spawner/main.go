// SPDX-FileCopyrightText: 2020 jecoz
//
// SPDX-License-Identifier: BSD-3-Clause

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/jecoz/flexi"
	"github.com/jecoz/flexi/fargate"
)

func exitf(format string, args ...interface{}) {
	fmt.Printf("error * "+format+"\n", args...)
	os.Exit(1)
}

func main() {
	var task flexi.Task
	if err := json.NewDecoder(os.Stdin).Decode(&task); err != nil {
		exitf("decode input task: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*3)
	defer cancel()

	rp, err := new(fargate.Fargate).Spawn(ctx, task)
	if err != nil {
		exitf("spawn task: %v", err)
	}

	if err = json.NewEncoder(os.Stdout).Encode(rp); err != nil {
		exitf("output encode: %v", err)
	}
}
