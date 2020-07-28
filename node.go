package go_data_routing

import "sync"

type nodeType int

//go:generate stringer -type=nodeType
const (
	source nodeType = iota + 1
	sink
	splitter
	processor
	to
	consumer

	filter
	aggregator
	mapper
	enricher
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
	- node owns input channel
	- output is a reference to consumer's input
	- node closes output channel
	- any given node especially "processor" should know if its output is consumed. unconsumed channel blocks the producer


	rt("bbb").process()
	rt("aaa").gen().enrich("bbb").split().sink()


	TODO
    - make poss. for a route to end with filter / processor

*/

type Exchange struct {
	ReqReply  bool
	CorrID    string
	Initiator *Node
	Msg       Job // both input & result
}

type NodeState struct {
	stopped bool
	typ     nodeType
	in      int
	isLast  bool
	err     error // error of runner
}

type Node struct {
	rt *Route

	input  chan Exchange
	output chan Exchange

	sync.Mutex
	NodeState

	runner func()
}

func (n *Node) onStop() {
	n.Lock()
	n.stopped = true
	n.Unlock()

	n.rt.onRunnerStop(n)
}
