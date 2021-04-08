package queuebroker

import (
	"sync"
	"sync/atomic"
)

// блокирующийся пул обработчиков
type WorkerPool struct {
	Handler func (interface{}) error
	wg *sync.WaitGroup // for gracefully shutdown
	done chan uint64
	tasks chan interface{}
	workers uint64
	relworkers uint64
	workwork int64
	sizeMux sync.Mutex
}

func (pool *WorkerPool) worker() {
	pool.wg.Add(1)
	defer pool.wg.Done()
	work := true
	for work {
		select {
		case t, ok := <- pool.tasks:
			if !ok {
				work = false
				break
			}
			atomic.AddInt64(&pool.workwork, 1)
			_ = pool.Handler(t) // TODO handle error
			atomic.AddInt64(&pool.workwork, -1)
		case <- pool.done:
			work = false
		}
	}
}

// блокирующая операция! Если ни один из обработчик не берется за обработку, то исполнение блокируется
func (pool *WorkerPool) Process(task interface{}) {
	pool.tasks <- task
}

// запускает пул обработчиков по количеству workers
func (pool *WorkerPool) Start(workers int) {
	pool.wg = &sync.WaitGroup{}
	pool.done = make(chan uint64)
	pool.tasks = make(chan interface{})
	pool.Grow(workers)
}

// увеличить число обработчиков на delta (строго положительное число)
func (pool *WorkerPool) Grow(delta int) {
	pool.sizeMux.Lock()
	defer pool.sizeMux.Unlock()
	for i := 0; i < delta; i++ {
		go pool.worker()
		atomic.AddUint64(&pool.workers, 1)
	}
}

// уменьшить число обработчиков на delta (строго положительное число)
func (pool *WorkerPool) Release(delta int) {
	pool.sizeMux.Lock()
	defer pool.sizeMux.Unlock()
	s := pool.Size()
	if delta > s {
		delta = s
	}
	for i := 0; i < delta; i++ {
		pool.done <- atomic.LoadUint64(&pool.workers)
	}
}

// количество активных обработчиков
func (pool *WorkerPool) Active() int {
	return int(atomic.LoadInt64(&pool.workwork))
}

// количество простаивающих обработчиков
func (pool *WorkerPool) Idle() int {
	return pool.Size() - pool.Active()
}

// возвращает количество обработчиков
func (pool *WorkerPool) Size() int {
	return int(atomic.LoadUint64(&pool.workers) - atomic.LoadUint64(&pool.relworkers))
}

// Ожидает завершение работы всех обработчиков
func (pool *WorkerPool) Shutdown() error {
	close(pool.done)
	pool.wg.Wait()
	return nil
}

// alias of Shutdown
func (pool *WorkerPool) Close() error {
	return pool.Shutdown()
}
