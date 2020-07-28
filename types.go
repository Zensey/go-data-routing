package go_data_routing

type Job interface {
	Run()
}

type Worker interface {
	SubmitJob(j Exchange)
}
