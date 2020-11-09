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

type Probe2 struct {
	x string
}

func (r *Probe2) Run() {
	time.Sleep(200 * time.Millisecond)
}

func TestEnrich(t *testing.T) {
	// TestRaceCondition
	for i := 0; i < 100; i++ {
		Enrich(t, 1*time.Second, 10*time.Millisecond)
	}

	// TestNormalMode
	for i := 0; i < 5; i++ {
		Enrich(t, 100*time.Millisecond, 2*time.Second)
	}

	EnrichTestOfDeadlock(t, 500*time.Millisecond, 10*time.Second)
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
			//fmt.Println("Sink >", e.Initiator)
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

func EnrichTestOfDeadlock(t *testing.T, sendMsgPeriod, cancelAfter time.Duration) {
	ctx, cancel := context.WithCancel(context.Background())
	rc := NewRouterContext(ctx)

	rc.Route("main").
		Source(func(n *Node) {
			for i := 0; i < 10; i++ {
				select {
				case i, _ := <-n.Input:
					if i.Type == Stop {
						return
					}

				case <-time.After(sendMsgPeriod):
					n.Send(Exchange{Msg: &Probe2{"a"}})
				}
			}
		}).
		To("enrich-rt").
		Sink(func(e Exchange) error {
			//fmt.Println("Sink >", e)
			return nil
		})

	rc.Route("enrich-rt").
		Process(1)

	go func() {
		time.Sleep(cancelAfter)
		cancel()
	}()

	rc.Run()
	rc.Print()

	// check all nodes are stopped by now
	for _, r := range rc.routes {
		for _, n := range r.nodes {
			if !n.stopped {
				t.Fail()
			}
		}
	}
}
