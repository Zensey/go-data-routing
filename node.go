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
	Type          exchangeType
	ReturnAddress chan Exchange
	Initiator     *Node
	//CorrID        string

	Header interface{}
	Msg    Job // both Input & result
}

type NodeState struct {
	stopped bool
	in      int
}

type Node struct {
	typ    nodeType
	isLast bool

	Input  chan Exchange
	Output chan Exchange

	sync.Mutex
	NodeState

	setup  func()
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

// Send an exchange to a next node
func (n *Node) Send(e Exchange) {
	//fmt.Printf("Send> %v +v | req-rep: %v, %v %v\n", n.isLast, e.Type == RequestReply, e.Initiator != nil, n.isLast)

	if n.isLast {
		// return reply to the initiator of Request-Reply
		if e.Type == RequestReply && e.Initiator != nil {
			e.ReturnAddress <- e
		}
		return
	}
	n.Output <- e
}
