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
						// Consider : using cancel ctx to term long-running requests ?
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

	var dstRoute *Route
	n.setup = func() {
		dstRoute = r.rc.lookupRoute(dst)
		dstRoute.grabReference()
	}
	n.runner = func() {
		dstNode := dstRoute.nodes.getFirstNode()
		countRequest := 0
		isClosing := false

		dstBuf := make(chan Exchange, 100)

		for {
			// 2 cases b/c of bufferisation (to ->dst) beeing used to prevent a deadlock

			var i Exchange
			if len(dstBuf) > 0 {
				i_ := <-dstBuf

				select {
				case i, _ = <-feedbackChan:
					dstBuf <- i_

				case dstNode.Input <- i_:
					continue
				}
			} else {
				select {
				case i, _ = <-n.Input:
				case i, _ = <-feedbackChan:
				}
			}

			if i.Type == Stop {
				isClosing = true
				if countRequest == 0 {
					dstRoute.releaseReference()
					return
				}

			} else {
				// detect type
				if i.Type == RequestReply && i.Initiator == n {

					// pass down
					i.Type = Request
					i.Initiator = nil
					n.Output <- i

					countRequest--
					if countRequest == 0 && isClosing {
						dstRoute.releaseReference()
						return
					}
				} else {
					n.incrIn()

					countRequest++
					i_ := i
					i_.Type = RequestReply
					i_.Initiator = n
					i_.ReturnAddress = feedbackChan

					dstBuf <- i_
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
