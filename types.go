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

/*
type RouterContext interface {
	Route(name string) *Route
}

// Route builder
type Route interface {
	Source(f func(n *Node)) *Route
	Filter(f func(e Exchange, n *Node)) *Route
	Process(nWorkers int) *Route
	To(dst string) *Route
	WireTap(dst string) *Route
	Sink(f func(e Exchange) error) *Route
}
*/
