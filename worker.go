package go_data_routing

type worker struct {
	pool    *BasicPool
	jobChan chan Exchange
}

func newWorker(p *BasicPool) *worker {
	return &worker{
		pool:    p,
		jobChan: make(chan Exchange),
	}
}

// Used by user-s code to submit a task to a worker
func (w *worker) SubmitJob(j Exchange) {
	w.jobChan <- j
}

// Used by pool to spawn a worker
func (w *worker) run(id int) {
	for {
		select {
		case w.pool.idleWorkers <- w:

		case <-w.pool.quit:
			w.pool.wg.Done()
			return
		}

		select {
		case job := <-w.jobChan:
			job.Msg.Run()
			w.pool.FuncOnJobResult(job)

		case <-w.pool.quit:
			w.pool.wg.Done()
			return
		}
	}
}
