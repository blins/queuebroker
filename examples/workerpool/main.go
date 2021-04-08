package main

import (
	"github.com/blins/queuebroker"
	"log"
	"math/rand"
	"time"
)

func main() {
	wp := queuebroker.WorkerPool{Handler: func(i interface{}) error {
		log.Printf("%v", i)
		<- time.After(100 * time.Millisecond)
		return nil
	}}

	wp.Start(5)
	defer wp.Close()

	go func () {
		for {
			log.Printf("Report: size=%v worked=%v idle=%v", wp.Size(), wp.Active(), wp.Idle())
			<- time.After(1 * time.Second)
		}
	}()
	tt := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < 300; i++ {
		wp.Process(i)
		<- time.After(time.Duration(tt.Intn(150)) * time.Millisecond)
	}
	log.Println("bye")
}
