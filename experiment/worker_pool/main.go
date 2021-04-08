package main

import (
	"log"
	"sync"
	"time"
)

func main() {

	worker := func(id int, wg *sync.WaitGroup, queue <- chan int, done <- chan int) {
		wg.Add(1)
		defer wg.Done()
		work := true
		for work {
			select {
			case i, ok := <- queue:
				if !ok {
					work = false
					break
				}
				log.Printf("ID: %v, Value: %v", id, i)
				<- time.After(1 * time.Second)
			case <- done:
				work = false
			}

		}
		log.Printf("ID: %v shutdown", id)
	}

	wg := &sync.WaitGroup{}
	ch := make(chan int)
	done := make(chan int)
	for j := 1; j < 6; j++ {
		go worker(j, wg, ch, done)
	}

	for j := 0; j < 30; j++ {
		log.Printf("send %v", j)
		ch <- j
	}
	close(done)
	wg.Wait()
	close(ch)
	log.Println("bye")
}
