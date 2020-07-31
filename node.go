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

/*
	Main idea:
	make a channel-based version to minimize memory usage
	stream processing instead of read all into array & then process: decrease latency and mem. consumption


	Architecture:

    route-1: [->N]-[->N]-...-[->N]
                   |
                   -other route

    Concerns / limitations:
	- each given node is connected with the next (if one exists) only with 1 channel
	- node owns Input channel
	- Output is a reference to consumer's Input
	- node does not close Output channel. Instead it sends msg type=close
	- any given node especially "processor" should know if its Output is consumed. unconsumed channel blocks the producer


	rt("bbb").process()
	rt("aaa").source().to("bbb").split().sink()


	TODO
    - make poss. for a route to end with filter / processor

*/

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
	isLast  bool
	err     error // error of runner
}

type Node struct {
	typ nodeType

	Input  chan Exchange
	Output chan Exchange

	sync.Mutex
	NodeState

	runner func()
}

func newNode(t nodeType) *Node {
	n := &Node{typ: t}

	if t != processor {
		n.Input = make(chan Exchange)
	}
	return n
}

func (n *Node) onStop() {
	n.Lock()
	n.stopped = true
	n.Unlock()
}

func (n *Node) incrIn() {
	n.Lock()
	n.in++
	n.Unlock()
}

func (n *Node) setErr(e error) {
	n.Lock()
	n.err = e
	n.Unlock()
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
