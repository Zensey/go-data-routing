package go_data_routing

import (
	"sync"
)

type Nodes []*Node

func (n *Nodes) getFirstNode() *Node {
	return (*n)[0]
}

type Route struct {
	rc    *RouterContext
	nodes Nodes

	wg sync.WaitGroup
}

func NewRouteBuilder(rc *RouterContext) *Route {
	r := &Route{
		rc:    rc,
		nodes: make(Nodes, 0),
	}
	return r
}

func (r *Route) grabReference() {
	r.wg.Add(1)
}

func (r *Route) releaseReference() {
	r.wg.Done()
}

func (r *Route) waitZeroReferences() {
	r.wg.Wait()
}

func (r *Route) addNode(n *Node) {
	r.nodes = append(r.nodes, n)
}

func (r *Route) isFirstNode(n *Node) bool {
	return r.nodes[0] == n
}

func (r *Route) Source(f func(n *Node)) *Route {
	n := newNode(source)
	n.runner = func() {
		f(n)
	}
	r.addNode(n)
	return r
}

func (r *Route) Filter(f func(j Exchange, n *Node)) *Route {
	n := newNode(filter)
	n.runner = func() {
		for {
			select {
			case i, _ := <-n.Input:
				if i.Type == Stop {
					return
				}

				n.incrIn()
				f(i, n)
			}
		}
	}
	r.addNode(n)
	return r
}

func (r *Route) Process(nWorkers int) *Route {
	n := newNode(processor)
	p := NewPluggablePool(nWorkers, n)
	p.setInputChan(n.Input)

	n.runner = func() {
		p.spawnWorkers()

		// exits from the cycle only when there's a spare worker and the job has been submitted to it
		for {
		l:
			select {
			case w := <-p.idleWorkers:
				select {
				case ex, _ := <-p.input:
					if ex.Type == Stop {
						// consider using cancel ctx to term long-running requests
						p.joinWorkers()
						return
					} else {
						n.incrIn()
						w.SubmitJob(ex)
						break l
					}
				}
			}
		}
	}
	r.addNode(n)
	return r
}

func (r *Route) To(dst string) *Route {
	n := newNode(to)
	feedbackChan := make(chan Exchange, 1)

	var destRoute *Route
	n.setup = func() {
		destRoute = r.rc.lookupRoute(dst)
		destRoute.grabReference()
	}

	n.runner = func() {
		destNode := destRoute.nodes.getFirstNode()
		countRequest := 0
		isClosing := false

		// buffer for exchanges being sent to the destination route
		dstBuf := make(chan Exchange, 100)
		for {
			var in Exchange

			// 2 cases b/c of bufferisation (to ->dst) being used to prevent a deadlock situation
			if len(dstBuf) > 0 {
				dst := <-dstBuf

				select {
				case in, _ = <-feedbackChan:
					dstBuf <- dst

				case destNode.Input <- dst:
					continue
				}
			} else {
				select {
				case in, _ = <-n.Input:
				case in, _ = <-feedbackChan:
				}
			}

			if in.Type == Stop {
				isClosing = true
				if countRequest == 0 {
					destRoute.releaseReference()
					return
				}

			} else {
				// detect type
				if in.Type == RequestReply && in.Initiator == n {

					// pass the exchange to the next node
					in.Type = Request
					in.Initiator = nil
					n.Output <- in

					countRequest--
					if countRequest == 0 && isClosing {
						destRoute.releaseReference()
						return
					}
				} else {
					n.incrIn()

					countRequest++
					copy := in
					copy.Type = RequestReply
					copy.Initiator = n
					copy.ReturnAddress = feedbackChan

					dstBuf <- copy
				}
			}
		}
	}
	r.addNode(n)
	return r
}

func (r *Route) WireTap(dst string) *Route {
	n := newNode(wiretap)

	var dstRoute *Route
	n.setup = func() {
		dstRoute = r.rc.lookupRoute(dst)
		dstRoute.grabReference()
	}
	n.runner = func() {
		dstNode := dstRoute.nodes.getFirstNode()

		for {
			select {
			case i, _ := <-n.Input:
				if i.Type == Stop {
					dstRoute.releaseReference()
					return
				}

				n.incrIn()
				if !dstNode.stopped {
					dstNode.Input <- i
				}
				n.Output <- i
			}
		}
	}

	r.addNode(n)
	return r
}

func (r *Route) Sink(f func(j Exchange) error) *Route {
	n := newNode(sink)
	n.runner = func() {
		for {
			select {
			case i, _ := <-n.Input:
				if i.Type == Stop {
					return
				}

				n.incrIn()
				f(i)

				if i.Type == RequestReply && i.Initiator != nil {
					// return to initiator
					i.ReturnAddress <- i
				}

			}
		}
	}
	r.addNode(n)
	return r
}
