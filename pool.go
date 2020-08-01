package go_data_routing

import (
	"sync"
)

/*
 Logic of pool

 1. wait for an idle worker
 2. wait for an incoming job
 3. on chan. close Quit all workers gracefully

*/

type PluggablePool struct {
	workers     []Worker
	idleWorkers chan Worker
	wg          sync.WaitGroup
	quit        chan bool
	input       chan Exchange
	results     chan Exchange
	nWorkers    int
	node        *Node
}

func NewPluggablePool(nWorkers int, n *Node) *PluggablePool {
	p := &PluggablePool{
		idleWorkers: make(chan Worker, 100),
		wg:          sync.WaitGroup{},
		quit:        make(chan bool),
		//input:       make(chan Exchange),
		nWorkers: nWorkers,
		node:     n,
	}

	return p
}

func (p *PluggablePool) Quit() chan bool {
	return p.quit
}

func (p *PluggablePool) IdleWorkers() chan Worker {
	return p.idleWorkers
}

func (p *PluggablePool) WorkerDone() {
	p.wg.Done()
}

func (p *PluggablePool) spawnWorker() {
	p.wg.Add(1)
	w := newWorker(p)
	go w.run()
}

func (p *PluggablePool) joinWorkers() {
	close(p.quit)
	p.wg.Wait()
}

func (p *PluggablePool) FuncOnJobResult(j Exchange) {
	p.node.Send(j)
}

func (p *PluggablePool) spawnWorkers() {
	for i := 0; i < p.nWorkers; i++ {
		p.spawnWorker()
	}
}

func (p *PluggablePool) setInputChan(i chan Exchange) {
	p.input = i
}
