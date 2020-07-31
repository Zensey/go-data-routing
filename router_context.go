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
	routes      map[string]*Nodes
	routesOrder []string
}

func NewRouterContext(ctx context.Context) *RouterContext {
	c := &RouterContext{
		ctx:    ctx,
		routes: make(map[string]*Nodes),
	}
	return c
}

func (c *RouterContext) Route(name string) *Route {
	c.routesWg.Add(1)
	rb := NewRouteBuilder(c)
	c.routes[name] = &rb.rt
	c.routesOrder = append(c.routesOrder, name)

	return rb
}

func (c *RouterContext) lookupRoute(s string) *Nodes {
	return c.routes[s]
}

func (c *RouterContext) Run() {
	fmt.Println("Routes Run > ")

	// link nodes by channels
	for _, rt := range c.routes {
		prev := (*Node)(nil)
		for i, n := range *rt {
			if prev != nil {
				prev.Output = n.Input
			}
			prev = n
			if i == len(*rt)-1 {
				n.isLast = true
			}
		}
	}

	// launch runners
	for _, rt := range c.routes {
		for _, n := range *rt {
			go func(n *Node) {
				n.runner()
				c.onRunnerStop(n)
			}(n)
		}
	}

	go func() {
		c.Print()
		for {
			select {
			case <-c.ctx.Done():
				for _, rn := range c.routesOrder {
					r := c.routes[rn]
					r.getFirstNode().Input <- Exchange{Type: Stop}
				}
				return

			case <-time.After(1 * time.Second):
				c.Print()
			}
		}
	}()

	c.routesWg.Wait() // wait for all routes
}

func (c *RouterContext) Print() {
	fmt.Print("\033[H\033[2J") // clear screen

	for _, rn := range c.routesOrder {
		r := c.routes[rn]

		fmt.Printf("Route: %s\n", rn)
		fmt.Println("type          in     Stop err")
		for _, n_ := range *r {
			n_.Lock()
			n := n_.NodeState
			n_.Unlock()

			fmt.Printf("└%-12s %-6d %v    %v\n", n_.typ.String(), n.in, boolToYN(n.stopped), n.err)
		}
	}
}

func (c *RouterContext) onRunnerStop(n *Node) {
	n.onStop()
	if !n.isLast {
		n.Send(Exchange{Type: Stop})
		return
	}
	c.routesWg.Done()
}

func boolToYN(b bool) string {
	if b {
		return "Y"
	}
	return "N"
}
