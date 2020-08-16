#### Data routing engine

Based on the concept of EIP (enterprise integration patterns) and go-pool-of-workers

The library provides the following primitives:
- route (chain of nodes processing messages)
- node:
    * _filter_
    * _processor_ -- processes a stream of tasks in parallel
    * _wire tap_: sends a copy of msg to another route (referenced by name)
    * _to_: enrich msg on another route (request-reply / enrichment pattern)

All the primitives are accessible through DSL.


Design of node:
- each node is connected with the next one (if exists) only with 1 channel
- node owns an _input_ channel
- _output_ is just a reference to the _input_ of next node
- node does not close the _output_ channel, instead it just sends a _Stop_ msg to a next node
- if a node is the last in a chain than an output message being sent is discarded unless it's not a _RequestReply_

Route builder DSL:
```cgo
type Route interface {
	Source (f func(n *Node)) Route
	Filter (f func(e Exchange, n *Node)) Route
	Process (workers int) Route
	To (route string) Route
	WireTap (route string) Route
	Sink (f func(e Exchange) error) Route
}
```

Examples:
- a simplistic bfs crawler: `go run examples/crawler/main.go`


#### Related links
* https://github.com/zensey/go-pool-of-workers
* https://github.com/chuntaojun/go-fork-join
* https://github.com/spatially/go-workgroup
