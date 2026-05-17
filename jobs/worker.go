package jobs

import "log"

// Job is the interface all background jobs must implement.
type Job interface {
	Process()
}

// Dispatcher manages a buffered channel queue and a pool of goroutine workers.
type Dispatcher struct {
	queue chan Job
}

// NewDispatcher creates a dispatcher with the given channel buffer size
// and starts the worker goroutine.
func NewDispatcher(bufferSize int) *Dispatcher {
	d := &Dispatcher{
		queue: make(chan Job, bufferSize),
	}
	go d.run()
	return d
}

// run listens on the queue channel and spawns a goroutine per job.
func (d *Dispatcher) run() {
	for job := range d.queue {
		go func(j Job) {
			defer func() {
				if r := recover(); r != nil {
					log.Printf("[JOB] Recovered from panic: %v", r)
				}
			}()
			j.Process()
		}(job)
	}
}

// Dispatch sends a job to the queue (non-blocking up to buffer size).
func (d *Dispatcher) Dispatch(job Job) {
	d.queue <- job
}
