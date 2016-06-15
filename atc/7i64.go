package main

import (
	"log"
	"time"

	"github.com/tarm/serial"
)

func newDevice() device {
	c := &serial.Config{Name: dev, Baud: baud}
	s, err := serial.OpenPort(c)
	if err != nil {
		log.Fatal(err)
	}

	mesa := device{
		name:        "7i64",
		port:        s,
		manCommands: make(chan string, 1),
	}

	_, err = s.Write([]byte{0x66, 0x00, 0x00, 0x00, 0x00, 0x00, 0x08})
	if err != nil {
		log.Println("Write Error")
	}
	time.Sleep(time.Millisecond * 200)
	return mesa
}

func (d *device) readIO() (IO, error) {
	d.Lock()
	defer d.Unlock()
	inbuf := make([]byte, 8)
	outbuf := make([]byte, 8)
	var io IO

	// Read Inputs
	for n := 0; n != 4; {
		d.port.Flush()
		_, err := d.port.Write([]byte{0x46, 0x04, 0x00})
		if err != nil {
			log.Println("Command Error")
		}
		time.Sleep(time.Millisecond * 3)
		n, err = d.port.Read(inbuf)
	}
	io.inputs = input(inbuf[3])<<24 + input(inbuf[2])<<16 + input(inbuf[1])<<8 + input(inbuf[0])

	// Read Outputs
	for n := 0; n != 4; {
		d.port.Flush()
		_, err := d.port.Write([]byte{0x46, 0x00, 0x00})
		if err != nil {
			log.Println("Command Error")
		}
		time.Sleep(time.Millisecond * 3)
		n, err = d.port.Read(outbuf)
	}
	io.outputs = output(outbuf[3])<<24 + output(outbuf[2])<<16 + output(outbuf[1])<<8 + output(outbuf[0])

	// Update Watchdog
	_, err := d.port.Write([]byte{0x66, 0x00, 0x00, outbuf[0], outbuf[1], outbuf[2], 0x08})
	time.Sleep(time.Millisecond * 3)
	if err != nil {
		log.Println(err)
	}

	ioChan <- io
	return io, nil
}

func (d *device) writeIO(pin output, on bool) error {

	d.Lock()
	defer d.Unlock()
	outbuf := make([]byte, 8)

	// Read Outputs
	for n := 0; n != 4; {
		d.port.Flush()
		_, err := d.port.Write([]byte{0x46, 0x00, 0x00})
		if err != nil {
			log.Println("Command Error")
		}
		time.Sleep(time.Millisecond * 3)
		n, err = d.port.Read(outbuf)
	}

	outputs := output(outbuf[3])<<24 + output(outbuf[2])<<16 + output(outbuf[1])<<8 + output(outbuf[0])

	if on {
		outputs |= pin
	} else {
		outputs &= ^pin
	}

	for i := 0; i < 4; i++ {
		outbuf[i] = byte(outputs >> uint8(i*8))
	}

	_, err := d.port.Write([]byte{0x66, 0x00, 0x00, outbuf[0], outbuf[1], outbuf[2], 0x08})

	return err
}
