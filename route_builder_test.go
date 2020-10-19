package go_data_routing

import (
	"context"
	"testing"
	"time"
)

type Probe struct {
	x string
}

func (r *Probe) Run() {}

func TestEnrich(t *testing.T) {

	// TestRaceCondition
	for i := 0; i < 100; i++ {
		Enrich(t, 1*time.Second, 10*time.Millisecond)
	}
	// TestNormalMode
	for i := 0; i < 5; i++ {
		Enrich(t, 100*time.Millisecond, 2*time.Second)
	}
}

func Enrich(t *testing.T, sendMsgPeriod, cancelAfter time.Duration) {
	ctx, cancel := context.WithCancel(context.Background())
	rc := NewRouterContext(ctx)

	rc.Route("main").
		Source(func(n *Node) {
			select {
			case i, _ := <-n.Input:
				if i.Type == Stop {
					return
				}

			case <-time.After(sendMsgPeriod):
				n.Send(Exchange{Msg: &Probe{"a"}})
			}
		}).
		To("enrich-rt").
		Sink(func(e Exchange) error {
			//fmt.Println("Sink >", e.Initiator.typ.String())
			return nil
		})

	rc.Route("enrich-rt").
		Process(1)

	go func() {
		time.Sleep(cancelAfter)
		cancel()
	}()
	rc.Run()

	// check all nodes are stopped by now
	for _, r := range rc.routes {
		for _, n := range r.nodes {
			if !n.stopped {
				t.Fail()
			}
		}
	}
}
