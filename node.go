package go_data_routing

import (
	"sync"
)

type nodeType int

//go:generate stringer -type=nodeType
const (
	source nodeType = iota + 1
	sink
	filter
	processor
	to
	wiretap
	consumer

	//aggregator
	//enricher
)

type exchangeType int

const (
	Request exchangeType = iota + 1
	RequestReply
	Stop
)

type Exchange struct {
	Type      exchangeType
	Initiator *Node
	//CorrID    string

	Header interface{}
	Msg    Job // both Input & result
}

type NodeState struct {
	stopped bool
	in      int
	err     error // error of runner
}

type Node struct {
	typ    nodeType
	isLast bool

	Input  chan Exchange
	Output chan Exchange

	sync.Mutex
	NodeState

	runner func()
}

func newNode(t nodeType) *Node {
	n := &Node{typ: t}
	n.Input = make(chan Exchange)

	return n
}

func (n *Node) onStop() {
	n.Lock()
	defer n.Unlock()
	n.stopped = true
}

func (n *Node) getStopped() bool {
	n.Lock()
	defer n.Unlock()
	return n.stopped
}

func (n *Node) getState() NodeState {
	n.Lock()
	defer n.Unlock()
	return n.NodeState
}

func (n *Node) incrIn() {
	n.Lock()
	defer n.Unlock()
	n.in++
}

func (n *Node) setErr(e error) {
	n.Lock()
	defer n.Unlock()
	n.err = e
}

func (n *Node) Send(e Exchange) {
	//fmt.Printf("Send> %v %+v\n", n.isLast, e)

	if n.isLast {
		// return reply to the initiator of Request-Reply
		if e.Type == RequestReply && e.Initiator != nil {
			e.Initiator.Input <- e
		}
		return
	}
	n.Output <- e
}
