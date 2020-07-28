package go_data_routing

import "sync"

type FuncOnJobResult func(Exchange)

type BasicPool struct {
	workers         []Worker
	idleWorkers     chan Worker
	wg              sync.WaitGroup
	quit            chan bool
	FuncOnJobResult FuncOnJobResult
	//results     chan Job
}

func NewBasicPool() BasicPool {
	return BasicPool{
		workers:     make([]Worker, 0),
		idleWorkers: make(chan Worker, 100),
		wg:          sync.WaitGroup{},
		quit:        make(chan bool),

		//results:     make(chan Job),
	}
}
