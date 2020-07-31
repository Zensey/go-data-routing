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

func Test_Enrich(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	rc := NewRouterContext(ctx)

	rc.Route("main").
		Source(func(n *Node) {
			for {
				select {
				case i, _ := <-n.Input:
					if i.Type == Stop {
						return
					}

				case <-time.After(1 * time.Second):
					n.Send(Exchange{Msg: &Probe{"a"}})
				}
			}
		}).
		To("enrich-rt").
		Sink(func(j_ Exchange) error {
			//fmt.Printf("Sink >> %+v \n", j_)
			return nil
		})

	rc.Route("enrich-rt").
		Process(1)

	go func() {
		time.Sleep(2 * time.Second)
		cancel()
		rc.Print()
	}()
	rc.Run()
	rc.Print()
}
