package go_data_routing

type Nodes []*Node

func (n *Nodes) getFirstNode() *Node {
	return (*n)[0]
}

type Route struct {
	rc *RouterContext
	rt Nodes
}

func NewRouteBuilder(rc *RouterContext) *Route {
	r := &Route{
		rc: rc,
		rt: make(Nodes, 0),
	}
	return r
}

func (r *Route) addNode(n *Node) {
	r.rt = append(r.rt, n)
}

func (r *Route) isFirstNode(n *Node) bool {
	return r.rt[0] == n
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
	n.Input = p.GetInputChan()

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
	n.runner = func() {
		dstRoute := r.rc.lookupRoute(dst)
		dstNode := dstRoute.getFirstNode()
		countTo := 0
		isClosing := false

		for {
			select {
			case i, _ := <-n.Input:
				if i.Type == Stop {
					isClosing = true
					if countTo == 0 {
						return
					}

				} else {
					// detect type
					if i.Type == RequestReply && i.Initiator == n {
						countTo--
						if countTo == 0 && isClosing {
							return
						}

						// pass down
						n.Output <- i
					} else {
						n.incrIn()

						countTo++
						i_ := i
						i_.Type = RequestReply
						i_.Initiator = n
						dstNode.Input <- i_
					}
				}
			}
		}
	}
	r.addNode(n)
	return r
}

func (r *Route) WireTap(dst string) *Route {
	n := newNode(wiretap)
	n.runner = func() {
		dstRoute := r.rc.lookupRoute(dst)
		dstNode := dstRoute.getFirstNode()
		for {
			select {
			case i, _ := <-n.Input:
				if i.Type == Stop {
					return
				}

				n.incrIn()
				if !dstNode.stopped { // possible data race. trace route dependencies / exchanges ?
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
				err := f(i)
				n.setErr(err)
				if n.err != nil {
					return
				}

				if i.Type == RequestReply && i.Initiator != nil { // return to initiator
					i.Initiator.Input <- i
				}
			}
		}
	}
	r.addNode(n)
	return r
}
