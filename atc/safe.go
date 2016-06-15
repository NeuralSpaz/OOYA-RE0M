package main

import (
	"fmt"
	"log"
	"time"
)

func Safe(d *device) stateFn {
	// d.Lock()
	// if d.mode == "Manual" {
	// 	d.Unlock()
	// 	return Manual
	// }
	// d.Unlock()

	ioJson.Lock()
	ioJson.State = "Safe"
	ioJson.Unlock()

	io, err := d.readIO()
	if err != nil {
		log.Fatal(err)
	}

	for key, _ := range atcControls.Control {
		atcControls.Control[key].Isdisabled = true
	}
	atcControls.Control[1].Isdisabled = false
	atcControls.Control[0].Isdisabled = false

	// fmt.Printf("out: %32.32b\n", io.outputs)
	// fmt.Printf("in:  %32.32b\n", io.inputs)
	// switch {
	// case !SpindleToolClampLS.isOn(io):
	// 	return SafeClamp
	// case !RetractRapidLS.isOn(io):
	// 	return SafeRetract
	// case UnClampedLS.isOn(io):
	// 	return SafeUp
	// case !OrientPinInsertedLS.isOn(io):
	// 	return SafeIndex
	// default:
	// 	return Safe

	if !SpindleToolClampLS.isOn(io) || SpindleToolUnClamp.isOn(io) {
		return SafeToolClamp
	}
	if !RetractedLS.isOn(io) && RetractedLS.isOn(io) {
		return SafeRetract
	}
	if UnClampedLS.isOn(io) || !InLS.isOn(io) || UnclampPin.isOn(io) || In.isOn(io) || Out.isOn(io) {
		return SafeUp
	}
	if !OrientPinInsertedLS.isOn(io) {
		return SafeIndex
	}

	nextState := make(chan stateFn, 1)

	go func() {
		select {
		case <-time.After(time.Millisecond * 1):
			nextState <- Safe
		case cmd := <-d.manCommands:
			fmt.Println("Recived Command in Safe:", cmd)
			// atcControls.Control[0].Isdisabled = false
			// ManControls, err := json.Marshal(atcControls)
			// if err != nil {
			// 	fmt.Println(err)
			// }
			// sendMan(ManControls)
			if cmd == "Manual" {
				// for key, _ := range atcControls.Control {
				// 	atcControls.Control[key].Isdisabled = false
				// }
				// atcControls.Control[1].Isdisabled = false
				// ManControls, err := json.Marshal(atcControls)
				// if err != nil {
				// 	fmt.Println(err)
				// }
				// sendMan(ManControls)
				nextState <- Manual

			} else {
				nextState <- Safe
			}

		}
	}()

	select {
	case state := <-nextState:
		return state
	}

}

func SafeToolClamp(d *device) stateFn {
	log.Println("SafeToolClamp")
	ioJson.Lock()
	ioJson.State = "SafeToolClamp"
	ioJson.Unlock()
	io, err := d.readIO()
	if err != nil {
		log.Fatal(err)
	}
	nextState := make(chan stateFn, 1)

	if !SpindleToolClampLS.isOn(io) || SpindleToolUnClamp.isOn(io) {
		fmt.Println("Clamping Tool")
		ioJson.Lock()
		ioJson.Statusz = "Clamping Tool"
		ioJson.Unlock()
		err := d.writeIO(SpindleToolUnClamp, OFF)
		if err != nil {
			log.Fatal(err)
		}
		done := make(chan bool, 1)
		go func() {
			for {
				io, err = d.readIO()
				if err != nil {
					log.Fatal(err)
				}
				if SpindleToolClampLS.isOn(io) {
					done <- true
					break
				}
			}
		}()

		select {
		case <-time.After(UnClampTimeout):
			log.Fatal("Timeout on Tool Limit Switch")
		case <-done:
			fmt.Println("Spindle Tool Clamped")

		}
	}
	if SpindleToolClampLS.isOn(io) && !SpindleToolUnClamp.isOn(io) {
		nextState <- Safe
	}

	select {
	case state := <-nextState:
		return state
	}

}

func SafeRetract(d *device) stateFn {
	log.Println("SafeRetract")
	ioJson.Lock()
	ioJson.State = "SafeRetract"
	ioJson.Unlock()
	io, err := d.readIO()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("out: %32.32b\n", io.outputs)
	fmt.Printf("in:  %32.32b\n", io.inputs)
	return Safe
}

func SafeUp(d *device) stateFn {
	log.Println("SafeUp")
	ioJson.Lock()
	ioJson.State = "SafeUp"
	ioJson.Unlock()
	nextState := make(chan stateFn, 1)

	err := d.writeIO(In, OFF)
	if err != nil {
		log.Fatal(err)
	}

	err = d.writeIO(Out, OFF)
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		io, err := d.readIO()
		if err != nil {
			log.Fatal(err)
		}

		if !InLS.isOn(io) {
			if !UnClampedLS.isOn(io) || !UnclampPin.isOn(io) {
				fmt.Println("Unclamping")
				err := d.writeIO(UnclampPin, ON)
				if err != nil {
					log.Fatal(err)
				}
				done := make(chan bool, 1)
				go func() {
					for {
						io, err = d.readIO()
						if err != nil {
							log.Fatal(err)
						}
						if UnClampedLS.isOn(io) {
							done <- true
							break
						}
					}
				}()

				select {
				case <-time.After(UnClampTimeout):
					log.Fatal("Timeout on UnClamped Limit Switch")
				case <-done:
					fmt.Println("UP/DOWN Unclamped")
					nextState <- SafeUp
				}
			} else {
				err := d.writeIO(In, ON)
				if err != nil {
					log.Fatal(err)
				}
				done := make(chan bool, 1)
				go func() {
					for {
						io, err = d.readIO()
						if err != nil {
							log.Fatal(err)
						}
						if InLS.isOn(io) {
							err := d.writeIO(In, OFF)
							if err != nil {
								log.Fatal(err)
							}
							done <- true
							break
						}
					}
				}()

				select {
				case <-time.After(InTimeout):
					log.Fatal("Timeout on In Limit Switch")
				case <-done:
					fmt.Println("ATC is up")
					nextState <- SafeUp
				}
			}
		} else {
			fmt.Println("Fully Up")
			if UnClampedLS.isOn(io) || UnclampPin.isOn(io) {
				fmt.Println("Clamping")
				ioJson.Lock()
				ioJson.Statusz = "Clamping"
				ioJson.Unlock()
				err := d.writeIO(UnclampPin, OFF)
				if err != nil {
					log.Fatal(err)
				}
				done := make(chan bool, 1)
				go func() {
					for {
						io, err = d.readIO()
						if err != nil {
							log.Fatal(err)
						}
						if !UnClampedLS.isOn(io) {
							done <- true
							break
						}
					}
				}()

				select {
				case <-time.After(UnClampTimeout):
					log.Fatal("Timeout on UnClamped Limit Switch while Clamping")
				case <-done:
					fmt.Println("UP/DOWN Clamped")
					ioJson.Lock()
					ioJson.Statusz = "UP/DOWN Clamped"
					ioJson.Unlock()
					nextState <- Safe
				}
			}
			nextState <- Safe
		}
	}()

	select {
	case state := <-nextState:
		return state
	}

}

func SafeIndex(d *device) stateFn {
	log.Println("SafeIndex")
	ioJson.Lock()
	ioJson.State = "SafeIndex"
	ioJson.Unlock()
	nextState := make(chan stateFn, 1)

	io, err := d.readIO()
	if err != nil {
		log.Fatal(err)
	}

	if CarouselPositionSprocket.isOn(io) {
		if PositionIsValid(io) {
			err := d.writeIO(OrientPin, OFF)
			if err != nil {
				log.Fatal(err)
			}
			done := make(chan bool, 1)
			go func() {
				for {
					io, err = d.readIO()
					if err != nil {
						log.Fatal(err)
					}
					if OrientPinInsertedLS.isOn(io) {
						done <- true
						break
					}
				}
			}()

			select {
			case <-time.After(OrientPinInsertTimeout):
				log.Fatal("Timeout on Orient Pin Insert Limit Switch")
			case <-done:
				fmt.Println("Orient Pin Inserted")
				nextState <- Safe
			}
		} else {
			log.Fatal("Error Carousel is indexed but with Invalid Position")
		}
	} else {
		if OrientPinRetractedLS.isOn(io) && OrientPin.isOn(io) {
			if RotateForwardRapidLS.isOn(io) || !RotateForwardRapidLS.isOn(io) && !RotateReverseRapidLS.isOn(io) {
				err := d.writeIO(CarouselRotateForwardSlow, ON)
				if err != nil {
					log.Fatal(err)
				}
				done := make(chan bool, 1)
				go func() {
					for {
						io, err = d.readIO()
						if err != nil {
							log.Fatal(err)
						}
						if CarouselPositionSprocket.isOn(io) {
							err := d.writeIO(CarouselRotateForwardSlow, OFF)
							if err != nil {
								log.Fatal(err)
							}
							done <- true
							break
						}
					}
				}()
				select {
				case <-time.After(RotateForwardSlowTimeout):
					log.Fatal("Timeout on CarouselPositionSprocket going Forward")
				case <-done:
					fmt.Println("Orient Pin Inserted")
					nextState <- SafeIndex
				}

			}
			if RotateReverseRapidLS.isOn(io) {
				err := d.writeIO(CarouselRotateReverseSlow, ON)
				if err != nil {
					log.Fatal(err)
				}
				done := make(chan bool, 1)
				go func() {
					for {
						io, err = d.readIO()
						if err != nil {
							log.Fatal(err)
						}
						if CarouselPositionSprocket.isOn(io) {
							err := d.writeIO(CarouselRotateReverseSlow, OFF)
							if err != nil {
								log.Fatal(err)
							}
							done <- true
							break
						}
					}
				}()
				select {
				case <-time.After(RotateReverseSlowTimeout):
					log.Fatal("Timeout on CarouselPositionSprocket going Reverse")
				case <-done:
					fmt.Println("Orient Pin Inserted")
					nextState <- SafeIndex
				}
			}

		} else {
			fmt.Println(OrientPin.isOn(io))
			fmt.Printf("out: %32.32b\n", io.outputs)
			err := d.writeIO(OrientPin, ON)
			if err != nil {
				log.Fatal(err)
			}
			done := make(chan bool, 1)
			go func() {
				for {
					io, err = d.readIO()
					if err != nil {
						log.Fatal(err)
					}
					if OrientPinRetractedLS.isOn(io) {
						done <- true
						break
					}
				}
			}()

			select {
			case <-time.After(OrientPinRetractTimeout):
				log.Fatal("Timeout on Orient Pin Retract Limit Switch")
			case <-done:
				fmt.Println("Orient Pin Retracted")
				nextState <- SafeIndex
			}
		}
		// Carousel not inPosition

	}
	select {
	case state := <-nextState:
		return state
	}
	// return Safe
}
