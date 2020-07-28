package go_data_routing

import (
	"fmt"
)

type Nodes []*Node

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

func (r *Route) Source(f func(send func(Exchange))) *Route {
	n := &Node{rt: r}
	n.NodeState.typ = source
	r.addNode(n)

	n.runner = func() {
		var sendFn = func(j Exchange) {
			n.output <- j
		}

		f(sendFn)
		n.onStop()
		close(n.output)
	}
	return r
}

func (r *Route) Filter(f func(j Exchange, sendFn func(Exchange))) *Route {
	n := &Node{rt: r}
	n.NodeState.typ = splitter
	n.input = make(chan Exchange)
	r.addNode(n)

	n.runner = func() {
		var sendFn = func(j Exchange) {
			n.output <- j
		}

		for {
			select {
			case i, ok := <-n.input:
				if !ok {
					n.onStop()
					close(n.output)
					return
				}
				n.Lock()
				n.in++
				n.Unlock()

				f(i, sendFn)
			}
		}
	}
	return r
}

func (r *Route) Process(nWorkers int) *Route {
	n := &Node{rt: r}
	n.NodeState.typ = processor
	r.addNode(n)

	workerPool := NewPluggablePool(nWorkers, n)

	n.input = workerPool.GetInputChan()
	n.runner = func() {
		//var sendFn = func(j Exchange) {
		//	n.output <- j
		//}

		workerPool.Run()
		n.onStop()
	}
	return r
}

func (r *Route) To(to string) *Route {
	n := &Node{rt: r}
	n.NodeState.typ = to
	n.input = make(chan Exchange)
	r.addNode(n)

	n.runner = func() {
		rr := r.rc.lookupRoute(to)
		fmt.Println("To >", rr, (*rr)[0], (*rr)[0].input)

		for {
			select {
			case i, ok := <-n.input:
				if !ok {
					n.onStop()

					close(n.output)
					return

				} else {
					// detect type
					if i.ReqReply && i.Initiator == n {
						fmt.Println("To > Answer", i.Msg)

						// pass down
						n.output <- i
					} else {
						n.Lock()
						n.in++
						n.Unlock()

						i_ := i
						i_.ReqReply = true
						i_.Initiator = n
						(*rr)[0].input <- i_
					}
				}
			}
		}
	}
	return r
}

func (r *Route) Sink(f func(j Exchange) error) *Route {
	n := &Node{rt: r}
	n.NodeState.typ = sink
	n.input = make(chan Exchange)
	r.addNode(n)

	n.runner = func() {
		//fmt.Println("Sink >", n.isLast)
	l:
		for {
			select {
			case j, ok := <-n.input:
				if !ok {
					break l
				}

				n.Lock()
				n.in++
				n.Unlock()

				err := f(j)
				n.Lock()
				n.err = err
				n.Unlock()

				if n.err != nil {
					break l
				}

				// return to initiator
			}
		}
		fmt.Println("Sink > stop")

		n.onStop()
	}
	return r
}

/////// ?
func (r *Route) onRunnerStop(n *Node) {
	//if n.next == nil {
	//	r.wg.Done()
	//}

	//fmt.Println("onRunnerStop >", n)
	if n.output == nil {
		r.rc.wg.Done()
	}
}
