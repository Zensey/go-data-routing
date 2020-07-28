package go_data_routing

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type RouterContext struct {
	wg  sync.WaitGroup
	ctx context.Context

	r map[string]*Nodes
}

func NewRouterContext(ctx context.Context) *RouterContext {
	r := &RouterContext{
		ctx: ctx,
		r:   make(map[string]*Nodes),
	}
	return r
}

func (rc *RouterContext) Route(name string) *Route {
	rb := NewRouteBuilder(rc)
	rc.r[name] = &rb.rt
	return rb
}

func (r *RouterContext) lookupRoute(s string) *Nodes {
	return r.r[s]
}

func (r *RouterContext) Run() {
	nr := 1 // number of routes
	r.wg.Add(nr)

	// link channels
	for _, rt := range r.r {
		prev := (*Node)(nil)
		for i, n := range *rt {
			if prev != nil {
				prev.output = n.input
			}
			prev = n
			if n.typ == to {
				//n.output_ = r.rt2[0].input
				//r.rt2[len(r.rt2)-1].output = n.input_
			}

			if i == len(*rt)-1 {
				n.isLast = true
			}
		}
	}

	// launch runners
	for _, rt := range r.r {
		for _, n := range *rt {
			go n.runner()
		}
	}

	quit := make(chan struct{})
	go func() {
		for {
			select {
			case <-quit:
				return
			case <-time.After(2 * time.Second):
				//r.Print()
				//case n:= <-r.nodeStop:
			}
		}
	}()

	r.wg.Wait()
	close(quit)
}

func (c *RouterContext) Print() {
	//fmt.Print("\033[H\033[2J")

	for i, rt := range c.r {
		fmt.Printf("Route: %s   in     stoped\n", i)
		for _, n_ := range *rt {
			n_.Lock()
			n := n_.NodeState
			n_.Unlock()

			fmt.Printf("└%-12s %-6d %v %v\n", n.typ.String(), n.in, n.stopped, n.err)
		}
	}
}
