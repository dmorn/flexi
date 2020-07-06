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
	fmt.Printf("error * "+format, args...)
	os.Exit(1)
}

func main() {
	var rp flexi.RemoteProcess
	if err := json.NewDecoder(os.Stdin).Decode(&rp); err != nil {
		exitf("decode input remote process: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()
	if err := new(fargate.Fargate).Kill(ctx, &rp); err != nil {
		exitf("kill remote process: %v", err)
	}
}
