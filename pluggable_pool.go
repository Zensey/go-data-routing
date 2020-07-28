package go_data_routing

import "fmt"

// Pool with an input Job channel
type PluggablePool struct {
	BasicPool

	input   chan Exchange
	results chan Exchange

	nWorkers int
	node     *Node
}

func NewPluggablePool(nWorkers int, n *Node) *PluggablePool {
	p := &PluggablePool{
		BasicPool: NewBasicPool(),

		input: make(chan Exchange),
		//results:  make(chan Job),

		nWorkers: nWorkers,
		node:     n,
	}
	p.FuncOnJobResult = p.onJobResult
	return p
}

func (p *PluggablePool) spawnWorker() {
	p.wg.Add(1)
	w := newWorker(&p.BasicPool)
	p.workers = append(p.workers, w)

	workerID := len(p.workers)
	go w.run(workerID)
}

func (p *PluggablePool) joinWorkers() {
	close(p.quit)
	p.wg.Wait()

	if p.results != nil {
		close(p.results)
	}

	//fmt.Println("joinWorkers> 4")
}

/*
 Logic of pool

 1. wait for worker
 2. wait for incoming job
 3. on chan. close quit all workers gracefully

*/

func (p *PluggablePool) onJobResult(j Exchange) {
	//fmt.Println("onJobResult >", j.ReqReply)

	if p.results != nil {
		p.results <- j
	} else {
		// return reply to initiator
		if j.ReqReply && j.Initiator != nil {
			j.Initiator.input <- j
		}
	}
}

func (p *PluggablePool) Run() {
	if p.node != nil /*&& p.node.output != nil*/ {
		p.results = p.node.output
	}

	for i := 0; i < p.nWorkers; i++ {
		p.spawnWorker()
	}

	// exits from the cycle only when there's a spare worker and the job has been submitted to it
	for {
	l:
		select {
		case w := <-p.idleWorkers:

			select {
			case j, ok := <-p.input:
				if ok {
					p.node.Lock()
					p.node.in++
					p.node.Unlock()

					w.SubmitJob(j)
					break l
				} else {
					// Consider : using cancel ctx to term long-running requests ?

					p.joinWorkers()
					fmt.Println("PluggablePool p.input > exit >")
					return
				}
			}
		}
	}
}

func (p *PluggablePool) GetInputChan() chan Exchange {
	return p.input
}
