package main

import "log"

// Automatic mode
//
//
//
//

func Auto(d *device) stateFn {
	_, err := d.readIO()
	if err != nil {
		log.Println(err)
	}

	// Get Command
	cmd := <-Commands
	if cmd.Cmd != "Auto" {
		Commands <- cmd
		return Manual
	}

	// Select Tool

	return Auto
}
