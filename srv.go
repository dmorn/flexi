package flexi

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"
)

type Srv struct {
	MakeTransporter func() (Transporter, error)
	Handler         func(context.Context, *Msg)
}

// TransporterRetry tries indefinetly to make a transporter. It exits with
// an error only if the context is invalidated.
func (s *Srv) TransporterRetry(ctx context.Context) (Transporter, error) {
	for {
		log.Printf("*** opening server transport ***")
		t, err := s.MakeTransporter()
		if err == nil {
			return t, nil
		}
		d := time.Second * 3
		log.Printf("error * %v", err)
		log.Printf("opening transport in %v", d)

		select {
		case <-time.After(d):
		case <-ctx.Done():
			return nil, fmt.Errorf("opening transport: %w", err)
		}
	}
}

func (s *Srv) Run(ctx context.Context) error {
	var wg sync.WaitGroup
	t, err := s.TransporterRetry(ctx)
	if err != nil {
		return err
	}
	for {
		msg, err := t.Recv(ctx)
		if err != nil {
			log.Printf("recv error * %v", err)
			if errors.Is(err, context.Canceled) {
				wg.Wait()
				return err
			}
			// Unless error is a context invalidation, we should
			// always retry connecting.
			t, err = s.TransporterRetry(ctx)
			if err != nil {
				wg.Wait()
				return err
			}
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
			s.Handler(ctx, msg)
		}()
	}
}
