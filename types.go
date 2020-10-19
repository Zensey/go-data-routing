package go_data_routing

type Job interface {
	Run()
}

type Worker interface {
	SubmitJob(Exchange)
}

type Pool interface {
	FuncOnJobResult(Exchange)
	Quit() chan bool
	IdleWorkers() chan Worker
	WorkerDone()
}
