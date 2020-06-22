package main

import ( "context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/jecoz/flexi"
	"github.com/jecoz/flexi/rabbit"
)


func main() {
	// TODO: flags
	addr := "amqp://localhost:5568"
	queue := "tomare"

	srv := &flexi.Srv{
		MakeTransporter: func() (flexi.Transporter, error) {
			return rabbit.NewTransport(addr, queue)
		},
		Handler: func(ctx context.Context, msg *flexi.Msg) {
		},
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := srv.Run(ctx); err != nil {
		log.Printf("error *** %v", err)
	}
}

