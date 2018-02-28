package goest_worker

import "errors"

var ErrorJobDropped = errors.New("job is dropped")
var ErrorJobPanic = errors.New("job is panic")
var ErrorJobBind = errors.New("first argument is not job instance")
