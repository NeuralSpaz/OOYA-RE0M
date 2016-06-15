package main

import (
	"encoding/json"
	"fmt"
	"log"
	"time"
)

type controls struct {
	Control [16]control `json:"control"`
}

type control struct {
	Title       string `json:"title"`
	Cmd         string `json:"cmd"`
	Icon        string `json:"icon"`
	Isdisabled  bool   `json:"isdisabled"`
	Isclickable bool   `json:"isClickable"`
}

var atcControls controls = controls{
	[16]control{control{
		Title:       "Manual",
		Cmd:         "Manual",
		Icon:        "fa fa-hand-paper-o",
		Isdisabled:  true,
		Isclickable: true,
	}, control{
		Title:       "Safe",
		Cmd:         "Safe",
		Icon:        "fa fa fa-refresh",
		Isdisabled:  false,
		Isclickable: true,
	}, control{
		Title:       "Unclamp",
		Cmd:         "Unclamp",
		Icon:        "fa fa-unlock",
		Isdisabled:  true,
		Isclickable: true,
	}, control{
		Title:       "Clamp",
		Cmd:         "Clamp",
		Icon:        "fa fa-lock",
		Isdisabled:  true,
		Isclickable: true,
	}, control{
		Title:       "Up",
		Cmd:         "Up",
		Icon:        "fa fa-arrow-up",
		Isdisabled:  true,
		Isclickable: true,
	}, control{
		Title:       "Down",
		Cmd:         "Down",
		Icon:        "fa fa-arrow-down",
		Isdisabled:  true,
		Isclickable: true,
	}, control{
		Title:       "Advance",
		Cmd:         "Advance",
		Icon:        "fa fa-arrow-right",
		Isdisabled:  true,
		Isclickable: true,
	}, control{
		Title:       "Retract",
		Cmd:         "Retract",
		Icon:        "fa fa-arrow-left",
		Isdisabled:  true,
		Isclickable: true,
	}, control{
		Title:       "Orient Pin Retrat",
		Cmd:         "OrientPinRetrat",
		Icon:        "fa fa-level-up",
		Isdisabled:  true,
		Isclickable: true,
	}, control{
		Title:       "Orient Pin Insert",
		Cmd:         "OrientPinInsert",
		Icon:        "fa fa-level-down",
		Isdisabled:  true,
		Isclickable: true,
	}, control{
		Title:       "Rotate CW",
		Cmd:         "RotateCW",
		Icon:        "fa fa-repeat",
		Isdisabled:  true,
		Isclickable: true,
	}, control{
		Title:       "Rotate CCW",
		Cmd:         "RotateCCW",
		Icon:        "fa fa-undo",
		Isdisabled:  true,
		Isclickable: true,
	}, control{
		Title:       "Index CW",
		Cmd:         "IndexCW",
		Icon:        "fa fa-fast-forward",
		Isdisabled:  true,
		Isclickable: true,
	}, control{
		Title:       "Index CCW",
		Cmd:         "IndexCCW",
		Icon:        "fa fa-fast-backward",
		Isdisabled:  true,
		Isclickable: true,
	}, control{
		Title:       "Tool Unclamp",
		Cmd:         "ToolUnclamp",
		Icon:        "fa fa-chain-broken",
		Isdisabled:  true,
		Isclickable: true,
	}, control{
		Title:       "Tool Clamp",
		Cmd:         "ToolClamp",
		Icon:        "fa fa-link",
		Isdisabled:  true,
		Isclickable: true,
	},
	},
}

var indexFWDTimer *time.Timer
var indexREVTimer *time.Timer

var ManLoopCount int

func Manual(d *device) stateFn {
	fmt.Println("Manual")

	io, err := d.readIO()
	if err != nil {
		log.Fatal(err)
	}
	var lastATC controls
	lastATC = atcControls
	for key, _ := range atcControls.Control {
		atcControls.Control[key].Isdisabled = false
	}
	atcControls.Control[1].Isdisabled = true
	atcControls.Control[0].Isdisabled = false

	if !UnClampedLS.isOn(io) {
		atcControls.Control[2].Isdisabled = false
	} else {
		atcControls.Control[2].Isdisabled = true
	}

	if UnClampedLS.isOn(io) {
		atcControls.Control[3].Isdisabled = false
	} else {
		atcControls.Control[3].Isdisabled = true
	}

	if UnclampPin.isOn(io) && UnClampedLS.isOn(io) {
		atcControls.Control[4].Isdisabled = false
		atcControls.Control[5].Isdisabled = false
	} else {
		atcControls.Control[4].Isdisabled = true
		atcControls.Control[5].Isdisabled = true
	}

	if OrientPin.isOn(io) && OrientPinRetractedLS.isOn(io) {
		atcControls.Control[10].Isdisabled = false
		atcControls.Control[11].Isdisabled = false
	} else {
		atcControls.Control[10].Isdisabled = true
		atcControls.Control[11].Isdisabled = true
	}

	if PositionIsValid(io) && !OrientPinInsertedLS.isOn(io) {
		atcControls.Control[9].Isdisabled = false
	} else {
		atcControls.Control[9].Isdisabled = true
	}

	if !OrientPinRetractedLS.isOn(io) {
		atcControls.Control[8].Isdisabled = false
	} else {
		atcControls.Control[8].Isdisabled = true
	}

	if SpindleToolUnClampLS.isOn(io) {
		atcControls.Control[14].Isdisabled = true
	} else {
		atcControls.Control[14].Isdisabled = false
	}

	if SpindleToolClampLS.isOn(io) {
		atcControls.Control[15].Isdisabled = true
	} else {
		atcControls.Control[15].Isdisabled = false
	}

	atcControlsJSON, err := json.Marshal(atcControls)
	if err != nil {
		fmt.Println(err)
	}
	if lastATC != atcControls || ManLoopCount > 20 {
		sendMan(atcControlsJSON)
		ManLoopCount = 0
	}
	ManLoopCount++
	
	select {
	case cmd := <-d.manCommands:
		// return cmd
		fmt.Println(cmd)
		switch cmd {
		case "Manual":
			go sendLog("Manual", "Manual")
			return Manual
		case "Safe":
			go sendLog("Safe", "Manual")
			return Safe
		case "Unclamp":
			go sendLog("Unclamp", "Manual")
			return Unclamp
		case "Clamp":
			go sendLog("Clamp", "Manual")
			return Clamp
		case "Up":
			go sendLog("Up", "Manual")
			return Up
		case "Down":
			go sendLog("Down", "Manual")
			return Down
		case "Advance":
			go sendLog("Advance", "Manual")
			return Advance
		case "Retract":
			go sendLog("Retract", "Manual")
			return Retract
		case "OrientPinRetrat":
			go sendLog("OrientPinRetrat", "Manual")
			return OrientPinRetrat
		case "OrientPinInsert":
			go sendLog("OrientPinInsert", "Manual")
			return OrientPinInsert
		case "RotateCW":
			go sendLog("RotateCW", "Manual")
			return RotateCW
		case "RotateCCW":
			go sendLog("RotateCCW", "Manual")
			return RotateCCW
		case "IndexCW":
			go sendLog("IndexCW", "Manual")
			return IndexCW
		case "IndexCCW":
			go sendLog("IndexCCW", "Manual")
			return IndexCCW
		case "ToolUnclamp":
			go sendLog("ToolUnclamp", "Manual")
			return ToolUnclamp
		case "ToolClamp":
			go sendLog("ToolClamp", "Manual")
			return ToolClamp
		default:
			return Manual
		}
	case <-time.After(time.Millisecond * 10):
		return Manual
	}
}

func Unclamp(d *device) stateFn {
	CurrentState := "Unclamp"
	log.Println(CurrentState)

	select {
	case nextState := <-d.manCommands:
		log.Println("got a interput cmd:", nextState)
		if nextState != CurrentState {
			d.manCommands <- nextState
			return Manual
		}
	case <-time.After(time.Millisecond * 1):
		io, err := d.readIO()
		if err != nil {
			log.Println(err)
		}
		if UnclampPin.isOn(io) && !UnClampedLS.isOn(io) {
			log.Println("Waiting for Carosel Clamp to Release")
		}
		if UnclampPin.isOn(io) && UnClampedLS.isOn(io) {
			log.Println("Carosel is UpClamped")
			go sendLog("Manual", CurrentState)
			return Manual
		}
		if !UnclampPin.isOn(io) {
			log.Println("Turning On Unclamp")
			go sendLog("Turning On Unclamp", CurrentState)
			err := d.writeIO(UnclampPin, ON)
			if err != nil {
				log.Println(err)
			}
		}
	}

	return Unclamp
}

func Clamp(d *device) stateFn {
	CurrentState := "Clamp"
	log.Println(CurrentState)

	select {
	case nextState := <-d.manCommands:
		log.Println("got a interrupt cmd:", nextState)
		if nextState != CurrentState {
			d.manCommands <- nextState
			return Manual
		}
	case <-time.After(time.Millisecond * 1):
		io, err := d.readIO()
		if err != nil {
			log.Println(err)
		}

		if InLS.isOn(io) && UnclampPin.isOn(io) {
			log.Println("Turning Off UnClamp")
			go sendLog("Turning Off UnClamp", CurrentState)
			err := d.writeIO(UnclampPin, OFF)
			if err != nil {
				log.Println(err)
			}
		}

		if !UnclampPin.isOn(io) && UnClampedLS.isOn(io) {
			log.Println("Clamping")
		}

		if !UnclampPin.isOn(io) && !UnClampedLS.isOn(io) {
			log.Println("Clamped")
			go sendLog("Manual", CurrentState)
			return Manual
		}
	}

	return Clamp
}

func Up(d *device) stateFn {
	CurrentState := "Up"
	log.Println(CurrentState)

	select {
	case nextState := <-d.manCommands:
		log.Println("got a interput cmd:", nextState)
		log.Println("Turning Off Up")
		go sendLog("Turning Off Up", CurrentState)
		err := d.writeIO(In, OFF)
		if err != nil {
			log.Println(err)
		}
		if nextState != CurrentState {
			d.manCommands <- nextState
			return Manual
		}
		return Manual
	case <-time.After(time.Millisecond * 1):
		io, err := d.readIO()
		if err != nil {
			log.Println(err)
		}

		if InLS.isOn(io) && In.isOn(io) {
			log.Println("Turning Off Up")
			go sendLog("Turning Off Up", CurrentState)
			err := d.writeIO(In, OFF)
			if err != nil {
				log.Println(err)
			}
		}
		if InLS.isOn(io) && !In.isOn(io) {
			log.Println("Carolel is Up")
			go sendLog("Carolel is Up", CurrentState)
			return Manual
		}
		if !InLS.isOn(io) && !In.isOn(io) && UnClampedLS.isOn(io) && UnclampPin.isOn(io) {
			log.Println("Turning on Up")
			go sendLog("Turning on Up", CurrentState)
			err := d.writeIO(In, ON)
			if err != nil {
				log.Println(err)
			}
		}
		if !UnClampedLS.isOn(io) || !UnclampPin.isOn(io) {
			log.Println("Need to Unclamp before going UP")
			// go sendLog("Need to Unclamp before going UP", CurrentState)
		}
	}

	return Up
}
func Down(d *device) stateFn {
	CurrentState := "Down"
	log.Println(CurrentState)

	select {
	case nextState := <-d.manCommands:
		log.Println("got a interput cmd:", nextState)
		log.Println("Turning Off Down")
		go sendLog("Turning Off Down", CurrentState)
		err := d.writeIO(Out, OFF)
		if err != nil {
			log.Println(err)
		}
		if nextState != CurrentState {
			d.manCommands <- nextState
			return Manual
		}
		return Manual
	case <-time.After(time.Millisecond * 1):
		io, err := d.readIO()
		if err != nil {
			log.Println(err)
		}

		if OutLS.isOn(io) && Out.isOn(io) {
			log.Println("Turning Off Down")
			go sendLog("Turning Off Down", CurrentState)
			err := d.writeIO(Out, OFF)
			if err != nil {
				log.Println(err)
			}
		}
		if OutLS.isOn(io) && !Out.isOn(io) {
			log.Println("Carolel is Down")
			go sendLog("Carolel is Down", CurrentState)
			return Manual
		}
		if !OutLS.isOn(io) && !Out.isOn(io) && UnClampedLS.isOn(io) && UnclampPin.isOn(io) {
			log.Println("Turning on Down")
			go sendLog("Turning on Down", CurrentState)
			err := d.writeIO(Out, ON)
			if err != nil {
				log.Println(err)
			}
		}
		if !UnClampedLS.isOn(io) || !UnclampPin.isOn(io) {
			log.Println("Need to Unclamp before going Down")
			// go sendLog("Need to Unclamp before going Down", CurrentState)
		}
	}

	return Down
}
func Advance(d *device) stateFn {
	CurrentState := "Advance"
	log.Println(CurrentState)

	select {
	case nextState := <-d.manCommands:
		log.Println("got a interput cmd:", nextState)
		log.Println("Turning Off Advance")
		go sendLog("Turning Off Advance", CurrentState)
		err := d.writeIO(AdvanceSlow, OFF)
		if err != nil {
			log.Println(err)
		}
		if nextState != CurrentState {
			d.manCommands <- nextState
			return Manual
		}
		return Manual
	case <-time.After(time.Millisecond * 1):
		io, err := d.readIO()
		if err != nil {
			log.Println(err)
		}
		// Do Work Here
		// Turning On CarouselRotateForwardSlow
		// if AdvanceSlowLS.isOn(io) {
		// 	return Manual
		// }

		if !AdvanceSlow.isOn(io) && !AdvanceSlowLS.isOn(io) {
			log.Println("Turning On Advance")
			go sendLog("Turning On Advance", CurrentState)
			err := d.writeIO(AdvanceSlow, ON)
			if err != nil {
				log.Println(err)
			}
			return Advance
		}
		if AdvanceSlowLS.isOn(io) {
			log.Println("Turning On Advance")
			go sendLog("Turning On Advance", CurrentState)
			err := d.writeIO(AdvanceSlow, OFF)
			if err != nil {
				log.Println(err)
			}
			return Manual
		}
	}
	return Advance
}

func Retract(d *device) stateFn {
	CurrentState := "Retract"
	log.Println(CurrentState)

	select {
	case nextState := <-d.manCommands:
		log.Println("got a interput cmd:", nextState)
		log.Println("Turning Off Retract")
		go sendLog("Turning Off Retract", CurrentState)
		err := d.writeIO(RetractSlow, OFF)
		if err != nil {
			log.Println(err)
		}
		if nextState != CurrentState {
			d.manCommands <- nextState
			return Manual
		}
		return Manual
	case <-time.After(time.Millisecond * 1):
		io, err := d.readIO()
		if err != nil {
			log.Println(err)
		}
		// Do Work Here
		// Turning On CarouselRotateForwardSlow
		// if RetractSlowLS.isOn(io) {
		// 	return Manual
		// }

		if !RetractSlow.isOn(io) && !RetractSlowLS.isOn(io) {
			log.Println("Turning On Retract")
			go sendLog("Turning On Retract", CurrentState)
			err := d.writeIO(RetractSlow, ON)
			if err != nil {
				log.Println(err)
			}
			return Retract
		}
		if RetractSlowLS.isOn(io) {
			log.Println("Turning On Retract")
			go sendLog("Turning On Retract", CurrentState)
			err := d.writeIO(RetractSlow, OFF)
			if err != nil {
				log.Println(err)
			}
			return Manual
		}
	}
	return Retract
}

func OrientPinRetrat(d *device) stateFn {
	CurrentState := "OrientPinRetrat"
	log.Println(CurrentState)

	select {
	case nextState := <-d.manCommands:
		log.Println("got a interput cmd:", nextState)
		if nextState != CurrentState {
			d.manCommands <- nextState
			return Manual
		}
	case <-time.After(time.Millisecond * 1):
		io, err := d.readIO()
		if err != nil {
			log.Println(err)
		}
		// Pin Retracted
		if OrientPinRetractedLS.isOn(io) && OrientPin.isOn(io) {
			log.Println("Pin Retracted")
			return Manual
		}
		// Pin Inserted Ouput off
		if !OrientPin.isOn(io) {
			log.Println("Turning On Pin Retract")
			go sendLog("urning On Pin Retract", CurrentState)
			err := d.writeIO(OrientPin, ON)
			if err != nil {
				log.Println(err)
			}
		}
		// Pin retracting
		if OrientPin.isOn(io) && !OrientPinRetractedLS.isOn(io) {
			log.Println("Pin Retracting")
		}
	}

	return OrientPinRetrat
}

func OrientPinInsert(d *device) stateFn {
	CurrentState := "OrientPinInsert"
	log.Println(CurrentState)

	select {
	case nextState := <-d.manCommands:
		log.Println("got a interput cmd:", nextState)
		if nextState != CurrentState {
			d.manCommands <- nextState
			return Manual
		}
	case <-time.After(time.Millisecond * 1):
		io, err := d.readIO()
		if err != nil {
			log.Println(err)
		}
		// Pin Retracted
		if OrientPinInsertedLS.isOn(io) && !OrientPin.isOn(io) {
			log.Println("Pin Inserted")
			return Manual
		}
		// Pin Inserted Ouput on
		if OrientPin.isOn(io) && PositionIsValid(io) {
			log.Println("Turning On Pin Retract")
			go sendLog("urning On Pin Retract", CurrentState)
			err := d.writeIO(OrientPin, OFF)
			if err != nil {
				log.Println(err)
			}
		}
		// Pin retracting
		if !OrientPin.isOn(io) && !OrientPinInsertedLS.isOn(io) {
			log.Println("Pin Inserting")
		}
	}

	return OrientPinInsert
}

func RotateCW(d *device) stateFn {
	CurrentState := "RotateCW"
	log.Println(CurrentState)

	select {
	case nextState := <-d.manCommands:
		log.Println("got a interput cmd:", nextState)
		log.Println("Turning Off CarouselRotateForwardSlow")
		go sendLog("Turning Off CarouselRotateForwardSlow", CurrentState)
		err := d.writeIO(CarouselRotateForwardSlow, OFF)
		if err != nil {
			log.Println(err)
		}
		if nextState != CurrentState {
			d.manCommands <- nextState
			return Manual
		}
		return Manual
	case <-time.After(time.Millisecond * 1):
		io, err := d.readIO()
		if err != nil {
			log.Println(err)
		}
		// Do Work Here
		if !OrientPinRetractedLS.isOn(io) {
			log.Println("Cannot Rotate with Pin inserted")
			return Manual
		}
		// Turning On CarouselRotateForwardSlow
		if OrientPinRetractedLS.isOn(io) && !CarouselRotateForwardSlow.isOn(io) {
			log.Println("Turning On CarouselRotateForwardSlow")
			go sendLog("Turning On CarouselRotateForwardSlow", CurrentState)
			err := d.writeIO(CarouselRotateForwardSlow, ON)
			if err != nil {
				log.Println(err)
			}
		}
	}
	return RotateCW
}

func RotateCCW(d *device) stateFn {
	const CurrentState = "RotateCCW"
	log.Println(CurrentState)

	select {
	case nextState := <-d.manCommands:
		log.Println("got a interput cmd:", nextState)
		log.Println("Turning Off CarouselRotateReverseSlow")
		go sendLog("Turning Off CarouselRotateReverseSlow", CurrentState)
		err := d.writeIO(CarouselRotateReverseSlow, OFF)
		if err != nil {
			log.Println(err)
		}
		if nextState != CurrentState {
			d.manCommands <- nextState

			return Manual
		}
		return Manual
	case <-time.After(time.Millisecond * 1):
		io, err := d.readIO()
		if err != nil {
			log.Println(err)
		}
		// Do Work Here
		if !OrientPinRetractedLS.isOn(io) {
			log.Println("Cannot Rotate with Pin inserted")
			return Manual
		}
		// Turning On CarouselRotateReverseSlow
		if OrientPinRetractedLS.isOn(io) && !CarouselRotateReverseSlow.isOn(io) {
			log.Println("Turning On CarouselRotateReverseSlow")
			go sendLog("Turning On CarouselRotateReverseSlow", CurrentState)
			err := d.writeIO(CarouselRotateReverseSlow, ON)
			if err != nil {
				log.Println(err)
			}
		}

	}

	return RotateCCW
}

func IndexCW(d *device) stateFn {
	CurrentState := "IndexCW"
	log.Println(CurrentState)

	select {
	case nextState := <-d.manCommands:
		log.Println("got a interput cmd:", nextState)
		if nextState == "Indexed" {
			err := d.writeIO(CarouselRotateForwardSlow, OFF)
			if err != nil {
				log.Println(err)
			}
			err = d.writeIO(OrientPin, OFF)
			if err != nil {
				log.Println(err)
			}
			return Manual
		}
		if nextState != CurrentState {
			d.manCommands <- nextState
			err := d.writeIO(CarouselRotateForwardSlow, OFF)
			if err != nil {
				log.Println(err)
			}
			indexFWDTimer.Stop()
			return Manual
		}
	case <-time.After(time.Millisecond * 1):
		io, err := d.readIO()
		if err != nil {
			log.Println(err)
		}

		if !OrientPinRetractedLS.isOn(io) {
			if !OrientPin.isOn(io) {
				err := d.writeIO(OrientPin, ON)
				if err != nil {
					log.Println(err)
				}
			}
		}
		if OrientPinRetractedLS.isOn(io) && !CarouselRotateForwardSlow.isOn(io) {
			err := d.writeIO(CarouselRotateForwardSlow, ON)
			if err != nil {
				log.Println(err)
			}
			indexFWDTimer = time.NewTimer(time.Millisecond * 500)
			go func() {
				select {
				case <-time.After(time.Millisecond * 520):
					log.Println("Killed Timer?")
					return
				case <-indexFWDTimer.C:
					for {
						log.Println("in go loop")
						io, err := d.readIO()
						if err != nil {
							log.Println(err)
						}
						if OrientPinRetractedLS.isOn(io) && CarouselRotateForwardSlow.isOn(io) && PositionIsValid(io) {
							// err := d.writeIO(CarouselRotateForwardSlow, OFF)
							// if err != nil {
							// 	log.Println(err)
							// }
							d.manCommands <- "Indexed"
							return
						}
					}
				}
			}()

		}

	}
	return IndexCW
}

func IndexCCW(d *device) stateFn {
	CurrentState := "IndexCCW"
	log.Println(CurrentState)

	select {
	case nextState := <-d.manCommands:
		log.Println("got a interput cmd:", nextState)
		if nextState == "Indexed" {
			err := d.writeIO(CarouselRotateReverseSlow, OFF)
			if err != nil {
				log.Println(err)
			}
			err = d.writeIO(OrientPin, OFF)
			if err != nil {
				log.Println(err)
			}
			return Manual
		}
		if nextState != CurrentState {
			d.manCommands <- nextState
			err := d.writeIO(CarouselRotateReverseSlow, OFF)
			if err != nil {
				log.Println(err)
			}
			indexREVTimer.Stop()
			return Manual
		}
	case <-time.After(time.Millisecond * 1):
		io, err := d.readIO()
		if err != nil {
			log.Println(err)
		}

		if !OrientPinRetractedLS.isOn(io) {
			if !OrientPin.isOn(io) {
				err := d.writeIO(OrientPin, ON)
				if err != nil {
					log.Println(err)
				}
			}
		}
		if OrientPinRetractedLS.isOn(io) && !CarouselRotateReverseSlow.isOn(io) {
			err := d.writeIO(CarouselRotateReverseSlow, ON)
			if err != nil {
				log.Println(err)
			}
			indexREVTimer = time.NewTimer(time.Millisecond * 500)
			go func() {
				select {
				case <-time.After(time.Millisecond * 520):
					log.Println("Killed Timer?")
					return
				case <-indexREVTimer.C:
					for {
						log.Println("in go loop")
						io, err := d.readIO()
						if err != nil {
							log.Println(err)
						}
						if OrientPinRetractedLS.isOn(io) && CarouselRotateReverseSlow.isOn(io) && PositionIsValid(io) {
							// err := d.writeIO(CarouselRotateReverseSlow, OFF)
							// if err != nil {
							// 	log.Println(err)
							// }
							d.manCommands <- "Indexed"
							return
						}
					}
				}
			}()

		}

	}
	return IndexCCW
}

func ToolUnclamp(d *device) stateFn {
	CurrentState := "ToolUnclamp"
	log.Println(CurrentState)

	select {
	case nextState := <-d.manCommands:
		log.Println("got a interput cmd:", nextState)
		if nextState != CurrentState {
			d.manCommands <- nextState
			return Manual
		}
	case <-time.After(time.Millisecond * 1):
		io, err := d.readIO()
		if err != nil {
			log.Println(err)
		}
		// UnClamped
		if SpindleToolUnClamp.isOn(io) && SpindleToolUnClampLS.isOn(io) {
			log.Println("ToolUnClamped")
			go sendLog("ToolUnClamped", CurrentState)
			return Manual
		}
		// Turn On
		if !SpindleToolUnClamp.isOn(io) {
			log.Println("Turning On ToolUnclamp")
			go sendLog("Turning On ToolUnclamp", CurrentState)
			err := d.writeIO(SpindleToolUnClamp, ON)
			if err != nil {
				log.Println(err)
			}
		}
		// Clamping
		if SpindleToolUnClamp.isOn(io) && !SpindleToolUnClampLS.isOn(io) {
			log.Println("Tool UnClamping")
		}
	}

	return ToolUnclamp
}

func ToolClamp(d *device) stateFn {
	CurrentState := "ToolClamp"
	log.Println(CurrentState)

	select {
	case nextState := <-d.manCommands:
		log.Println("got a interput cmd:", nextState)
		if nextState != CurrentState {
			d.manCommands <- nextState
			return Manual
		}
	case <-time.After(time.Millisecond * 1):
		io, err := d.readIO()
		if err != nil {
			log.Println(err)
		}
		// Clamped
		if !SpindleToolUnClamp.isOn(io) && SpindleToolClampLS.isOn(io) {
			log.Println("ToolClamped")
			go sendLog("ToolClamped", CurrentState)
			return Manual
		}
		// Turn Off
		if SpindleToolUnClamp.isOn(io) {
			log.Println("Turning Off ToolUnclamp")
			go sendLog("Turning Off ToolUnclamp", CurrentState)
			err := d.writeIO(SpindleToolUnClamp, OFF)
			if err != nil {
				log.Println(err)
			}
		}
		// Clamping
		if !SpindleToolUnClamp.isOn(io) && !SpindleToolClampLS.isOn(io) {
			log.Println("Tool Clamping")
		}
	}

	return ToolClamp
}
