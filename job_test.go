package goest_worker

type mockJobCall struct {
	priority	int
}

func (job *mockJobCall) call() {

}


func (job *mockJobCall) Priority() int {
	return job.priority
}