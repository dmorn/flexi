package main

import (
	"log"

	"github.com/jecoz/flexi/fargate"
	"github.com/jecoz/flexi/styx"
)

func main() {
	srv := styx.NewSrv(fargate.New())
	if err := srv.Serve("9pfs"); err != nil {
		log.Printf("error * %v", err)
	}
}
