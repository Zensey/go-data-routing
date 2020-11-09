package go_data_routing

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type RouterContext struct {
	routesWg    sync.WaitGroup
	ctx         context.Context
	routes      map[string]*Route
	routesOrder []string
}

func NewRouterContext(ctx context.Context) *RouterContext {
	c := &RouterContext{
		ctx:    ctx,
		routes: make(map[string]*Route),
	}
	return c
}

func (c *RouterContext) Route(name string) *Route {
	c.routesWg.Add(1)
	rb := NewRouteBuilder(c)
	c.routes[name] = rb
	c.routesOrder = append(c.routesOrder, name)

	return rb
}

func (c *RouterContext) lookupRoute(s string) *Route {
	return c.routes[s]
}

func (c *RouterContext) Run() {
	fmt.Println("RouterContext > Run")

	// link nodes by channels
	for _, rt := range c.routes {
		prev := (*Node)(nil)
		for i, n := range rt.nodes {
			if n.setup != nil {
				n.setup()
			}

			if prev != nil {
				prev.Output = n.Input
			}
			prev = n
			if i == len(rt.nodes)-1 {
				n.isLast = true
			}
		}
	}

	// launch runners
	for _, rt := range c.routes {
		for _, n := range rt.nodes {
			go func(n *Node) {
				n.runner()
				c.onRunnerStop(n)
			}(n)
		}
	}

	go func() {
		//c.Print()
		for {
			select {
			case <-c.ctx.Done():
				fmt.Println("Stopping..")
				for _, rn := range c.routesOrder {
					r := c.routes[rn]
					r.waitZeroReferences()

					if !r.nodes.getFirstNode().getStopped() {
						r.nodes.getFirstNode().Input <- Exchange{Type: Stop}
					}
				}
				return

			case <-time.After(5 * time.Second):
				c.Print()
			}
		}
	}()

	c.routesWg.Wait() // wait for all routes to stop
}

func (c *RouterContext) Print() {
	//fmt.Print("\033[H\033[2J") // clear screen

	for _, rn := range c.routesOrder {
		r := c.routes[rn]

		fmt.Printf("Route: %s\n", rn)
		fmt.Println("type          in     Stop")
		for _, n := range r.nodes {
			s := n.getState()

			fmt.Printf("└%-12s %-6d %v      @ %p\n", n.typ.String(), s.in, boolToYN(s.stopped), n)
		}
	}
}

func (c *RouterContext) onRunnerStop(n *Node) {
	n.onStop()
	if !n.isLast {
		n.Send(Exchange{Type: Stop})
		return
	}
	// last node in a route
	c.routesWg.Done()
}

func boolToYN(b bool) string {
	if b {
		return "Y"
	}
	return "N"
}
