# go-data-routing

![Test](https://github.com/Zensey/go-data-routing/workflows/Test/badge.svg?branch=dev)


The library provides a pipeline engine for stream processing of data.
It is based on the concept of EIP (enterprise integration patterns).

The motivation for this project is simple: to get an easy and clear way of coding ETL-like programs for parallel processing of data. In my case it was a BFS crawler tuned for extraction of specific metadata, (see a basic version in `example` folder). 

# Features
The library provides the following primitives:
- route (chain of nodes processing messages)
- node:
    * _filter_ -- filter a stream of exchanges
    * _processor_ -- processes a stream of exchanges (tasks) in parallel
    * _wire tap_: sends a copy of msg to another route (referenced by name)
    * _to_: enrich msg on another route (request-reply / enrichment pattern)

All the primitives are accessible through DSL.

## Route builder DSL (methods):

Method | Signature | Args
--- | --- | ---
Source | f func(n *Node) | function f, used for generation of exchanges 
Filter | f func(e Exchange, n *Node) | function f, intercepts exchanges
Process| workers int | Number of workers, beeing used to process exchanges
To| route string | Name of the route, where to _redirect_ an exchange for execution of request-reply
WireTap| route string | Name of the route, where to _copy_ an exchange
Sink| f func(e Exchange) error | Function f, used for consumption exchanges


## Examples:
- a simplistic bfs crawler: `go run examples/crawler/main.go`


## Design of node:
- each node is connected with the next one (if exists) only with 1 channel
- node owns an _input_ channel
- _output_ is just a reference to the _input_ of next node
- node does not close the _output_ channel, instead it just sends a _Stop_ msg to a next node
- if a node is the last in a chain than an output message being sent is discarded unless it's not a _RequestReply_

## Code coverage
See the [coverage.html](coverage.html) for details.


## Related links
* [A Stream Processing API for Go](https://medium.com/@vladimirvivien/a-stream-processing-api-for-go-842676efe315)
* https://github.com/zensey/go-pool-of-workers
* https://github.com/chuntaojun/go-fork-join
* https://github.com/spatially/go-workgroup
