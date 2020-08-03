package go_data_routing

import (
	"context"
	"testing"
	"time"
)

type Probe struct {
	X string
}

func (r *Probe) Run() {}

func TestEnrich(t *testing.T) {

	for i := 0; i < 1000; i++ {
		Enrich(t)
	}
}

func Enrich(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	rc := NewRouterContext(ctx)

	rc.Route("main").
		Source(func(n *Node) {
			select {
			case i, _ := <-n.Input:
				if i.Type == Stop {
					return
				}

			case <-time.After(1 * time.Second):
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
		time.Sleep(10 * time.Millisecond)
		cancel()
		rc.Print()
	}()
	rc.Run()
	rc.Print()
	// check that all nodes are stoped
	for _, r := range rc.routes {
		for _, n := range *r {
			if !n.stopped {
				t.Fail()
			}
		}
	}

}
