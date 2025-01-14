package server

import (
	"fmt"
	"sync"
	"time"
)

/*
when a new task comes in
we check the waitingQueue ,if yes, then we add to the back of the queue, when no new task comes in but waiting queue has length then we remove task one by one from the waitin queue for execution
when no waitingQueue and a task comes in and the workers are available we execute immediately,if no worker available and worker length is less than max we assign immediately to the new worker created. but when more than max then add to waitingQueue
when workers are idle we remove them
*/

const (
	IDLE_TIMEOUT_SEC = 2 * time.Second
)

type Pool struct {
	maxWorker    int
	taskQueue    chan (func())
	workerQueue  chan (func())
	waitingQueue []func()
	stop         chan struct{}
	shutdown     bool
	lock         sync.Mutex
}

/**
 * NewWorkerPool is used to create a new worker pool
 * @param max int - the maximum number of workers
 * @return *Pool - the new worker pool
 */
func NewWorkerPool(max int) *Pool {
	if max < 1 {
		max = 1
	}
	newPool := &Pool{
		maxWorker:   max,
		taskQueue:   make(chan func()),
		workerQueue: make(chan func()),
		stop:        make(chan struct{}),
	}

	//fmt.Println("Worker pool created with max workers:", max)

	go newPool.start()

	return newPool
}

/**
 * Submit is used to submit a task to the worker pool
 * @param task func() - the task to submit
 */
func (p *Pool) Submit(task func()) {
	if task == nil {
		return
	}
	p.lock.Lock()
	defer p.lock.Unlock()

	if p.shutdown {
		fmt.Println("Cannot submit task: worker pool is shutting down.")
		return
	}

	p.taskQueue <- task

}

/**
 * start is used to start the worker pool
 */
func (p *Pool) start() {
	var workerLen int
	var idle bool
	timeout := time.NewTimer(IDLE_TIMEOUT_SEC)
	var wg sync.WaitGroup

Loop:
	for {

		if len(p.waitingQueue) != 0 {
			if !p.processWaiting() {
				break Loop
			}
			continue
		}

		select {
		case task, ok := <-p.taskQueue:
			if !ok {
				break Loop
			}

			select {
			case p.workerQueue <- task:
			default:
				if workerLen < p.maxWorker {
					wg.Add(1)
					go worker(task, p.workerQueue, &wg)
					workerLen++
				} else {
					p.waitingQueue = append(p.waitingQueue, task)

				}
			}

			idle = false

		case <-timeout.C:
			// prevents premature removal
			if idle && workerLen > 0 {
				if p.killIdleWorker() {
					workerLen--
				}
			}
			idle = true
			timeout.Reset(IDLE_TIMEOUT_SEC)
		case <-p.stop:
			break Loop
		}

	}

	for len(p.waitingQueue) > 0 {
		p.workerQueue <- p.waitingQueue[0]
		p.waitingQueue = p.waitingQueue[1:]
	}
	for workerLen > 0 {
		p.workerQueue <- nil
		workerLen--
	}
	wg.Wait()
	timeout.Stop()

	close(p.workerQueue)
}

/**
 * worker is used to execute the task
 * @param task func() - the task to execute
 * @param workerQueue chan func() - the worker queue
 * @param wg *sync.WaitGroup - the wait group
 */
func worker(task func(), workerQueue chan func(), wg *sync.WaitGroup) {
	defer wg.Done()

	for task != nil {
		task()
		task = <-workerQueue
	}

}

/**
 * killIdleWorker is used to kill the idle worker
 * @return bool - true if the worker is killed, false otherwise
 */
func (p *Pool) killIdleWorker() bool {
	select {
	case p.workerQueue <- nil:
		return true
	default:
		return false
	}
}

/**
 * processWaiting is used to process the waiting queue
 * @return bool - true if the waiting queue is processed, false otherwise
 */
func (p *Pool) processWaiting() bool {
	select {
	case task, ok := <-p.taskQueue:
		if !ok {
			return false
		}
		p.waitingQueue = append(p.waitingQueue, task)

	case p.workerQueue <- p.waitingQueue[0]:
		p.waitingQueue = p.waitingQueue[1:]
	}

	return true
}

/**
 * Shutdown is used to shutdown the worker pool
 */
func (p *Pool) Shutdown() {
	p.lock.Lock()
	defer p.lock.Unlock()

	if p.shutdown {
		return
	}

	p.shutdown = true
	close(p.taskQueue)
	close(p.stop)
}
