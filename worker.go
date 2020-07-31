package go_data_routing

type worker struct {
	pool    Pool
	jobChan chan Exchange
}

func newWorker(p Pool) *worker {
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
func (w *worker) run() {
	for {
		select {
		case w.pool.IdleWorkers() <- w:

		case <-w.pool.Quit():
			w.pool.WorkerDone()
			return
		}

		select {
		case job := <-w.jobChan:
			job.Msg.Run()
			w.pool.FuncOnJobResult(job)

		case <-w.pool.Quit():
			w.pool.WorkerDone()
			return
		}
	}
}
