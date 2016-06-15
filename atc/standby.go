package main

import (
	"fmt"
	"log"
	"time"
)

func Standby(d *device) stateFn {
	log.Println("Standby")
	ioJson.Lock()
	ioJson.State = "Standby"
	ioJson.Unlock()
	nextState := make(chan stateFn, 1)
	errorChan := make(chan error, 1)

	go func() {
		io, err := d.readIO()
		if err != nil {
			errorChan <- err
		}
		fmt.Printf("out: %32.32b\n", io.outputs)
		fmt.Printf("in:  %32.32b\n", io.inputs)
		nextState <- Standby

	}()

	for {
		select {
		case <-time.After(time.Millisecond * 100):
			fmt.Println("TimedOut")
			return nil
		case state := <-nextState:
			return state
		case err := <-errorChan:
			log.Println(err)
			return nil
		}
	}
}
