package goest_worker

type WorkerInterface interface {
	// start proccess
	Start()

	// get quit channel for soft exit
	GetQuitChan() (chan bool)

	// add job to work
	AddJob(JobInstance) ()
}

