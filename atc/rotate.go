package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	_ "expvar"

	"github.com/gorilla/websocket"
	"github.com/tarm/serial"
)

// Inputs
type input uint32
type output uint32

const (
	OrientPinRetractedLS     input = 0x01 << 0
	OrientPinInsertedLS      input = 0x01 << 1
	RotateForwardRapidLS     input = 0x01 << 2
	RotateReverseRapidLS     input = 0x01 << 3
	AdvancedLS               input = 0x01 << 4
	RetractedLS              input = 0x01 << 5
	AdvanceSlowLS            input = 0x01 << 7
	RetractSlowLS            input = 0x01 << 6
	InLS                     input = 0x01 << 8
	OutLS                    input = 0x01 << 9
	SpindleToolUnClampLS     input = 0x01 << 11
	SpindleToolClampLS       input = 0x01 << 10
	SpindleGearLowLS         input = 0x01 << 12
	SpindleGearHighLS        input = 0x01 << 13
	UnClampedLS              input = 0x01 << 14
	CarouselPositionBit0     input = 0x01 << 16
	CarouselPositionBit1     input = 0x01 << 17
	CarouselPositionBit2     input = 0x01 << 18
	CarouselPositionBit3     input = 0x01 << 19
	CarouselPositionBit4     input = 0x01 << 20
	CarouselPositionBit5     input = 0x01 << 21
	CarouselPositionSprocket input = 0x01 << 22
)

var (
	OrientPinInsertTimeout   time.Duration = time.Millisecond * 10000
	OrientPinRetractTimeout  time.Duration = time.Millisecond * 10000
	InTimeout                time.Duration = time.Millisecond * 10000
	OutTimeout               time.Duration = time.Millisecond * 10000
	AdvanceTimeout           time.Duration = time.Millisecond * 10000
	RetractTimeout           time.Duration = time.Millisecond * 10000
	RotateForwardSlowTimeout time.Duration = time.Millisecond * 10000
	RotateReverseSlowTimeout time.Duration = time.Millisecond * 10000
	UnClampTimeout           time.Duration = time.Millisecond * 10000
	ClampTimeout             time.Duration = time.Millisecond * 10000
)

const (
	OrientPin                  output = 0x01 << 0
	CarouselRotateForwardRapid output = 0x01 << 1
	CarouselRotateReverseRapid output = 0x01 << 2
	CarouselRotateForwardSlow  output = 0x01 << 3
	CarouselRotateReverseSlow  output = 0x01 << 4
	AdvanceRapid               output = 0x01 << 5
	RetractRapid               output = 0x01 << 6
	AdvanceSlow                output = 0x01 << 7
	RetractSlow                output = 0x01 << 8
	In                         output = 0x01 << 9
	Out                        output = 0x01 << 10
	UnclampPin                 output = 0x01 << 11
	SpindleToolUnClamp         output = 0x01 << 12
	ZAxisClamp                 output = 0x01 << 13
	SpindleGearLow             output = 0x01 << 14
	SpindleGearHigh            output = 0x01 << 15
)

type ioStatus struct {
	sync.RWMutex
	State            string `json:"state"`
	Statusz          string `json:"statusz"`
	CarouselPosition int    `json:"carouselposition"`
	Intputs          struct {
		OrientPinRetractedLS     bool `json:"Orient Pin Retracted LS"`
		OrientPinInsertedLS      bool `json:"Orient Pin Inserted LS"`
		RotateForwardRapidLS     bool `json:"Rotate Forward Rapid LS"`
		RotateReverseRapidLS     bool `json:"Rotate Reverse Rapid LS"`
		AdvancedLS               bool `json:"Advanced LS"`
		RetractedLS              bool `json:"Retracted LS"`
		AdvanceSlowLS            bool `json:"Advance Slow LS"`
		RetractSlowLS            bool `json:"Retract Slow LS"`
		InLS                     bool `json:"In LS"`
		OutLS                    bool `json:"Out LS"`
		SpindleToolUnClampLS     bool `json:"Spindle Tool UnClamp LS"`
		SpindleToolClampLS       bool `json:"Spindle Tool Clamp LS"`
		SpindleGearLowLS         bool `json:"Spindle Gear Low LS"`
		SpindleGearHighLS        bool `json:"Spindle Gear High LS"`
		UnClampedLS              bool `json:"UnClamped LS"`
		CarouselPositionBit0     bool `json:"Carousel Position Bit 0"`
		CarouselPositionBit1     bool `json:"Carousel Position Bit 1"`
		CarouselPositionBit2     bool `json:"Carousel Position Bit 2"`
		CarouselPositionBit3     bool `json:"Carousel Position Bit 3"`
		CarouselPositionBit4     bool `json:"Carousel Position Bit 4"`
		CarouselPositionBit5     bool `json:"Carousel Position Bit 5"`
		CarouselPositionSprocket bool `json:"Carousel Position Sprocket"`
	} `json:"inputs"`
	Outputs struct {
		OrientPin                  bool `json:"Orient Pin"`
		CarouselRotateForwardRapid bool `json:"Carousel Rotate Forward Rapid"`
		CarouselRotateReverseRapid bool `json:"Carousel Rotate Reverse Rapid"`
		CarouselRotateForwardSlow  bool `json:"Carousel Rotate Forward Slow"`
		CarouselRotateReverseSlow  bool `json:"Carousel Rotate Reverse Slow"`
		AdvanceRapid               bool `json:"Advance Rapid"`
		RetractRapid               bool `json:"Retract Rapid"`
		AdvanceSlow                bool `json:"Advance Slow"`
		RetractSlow                bool `json:"Retract Slow"`
		In                         bool `json:"In"`
		Out                        bool `json:"Out"`
		UnclampPin                 bool `json:"Unclamp Pin"`
		SpindleToolUnClamp         bool `json:"Spindle Tool UnClamp"`
		SpindleToolClamp           bool `json:"Spindle Tool Clamp"`
		SpindleGearLow             bool `json:"Spindle Gear Low"`
		SpindleGearHigh            bool `json:"Spindle Gear High"`
		ZAxisClamp                 bool `json:"ZAxis Clamp"`
	} `json:"outputs"`
}

const (
	ON  bool = true
	OFF bool = false
)

var ErrFailedRead = errors.New("Failed io Read")

type device struct {
	sync.Mutex
	lastState stateFn
	mode      string
	name      string
	version   string
	port      *serial.Port
}

type IO struct {
	inputs  input
	outputs output
}

var dev string
var baud int
var listen = flag.String("Listen", "0.0.0.0:1234", "host:port")

var (
	Trace   *log.Logger
	Info    *log.Logger
	Warning *log.Logger
	Error   *log.Logger
)

func init() {
	flag.StringVar(&dev, "dev", "/dev/ttyUSB0", "7i64 USB PORT")
	flag.IntVar(&baud, "baud", 115200, "Baud Rate")

	// logwriter, e := syslog.New(syslog.LOG_NOTICE, "Bureau of Meteorology")
	// if e == nil {
	// 	log.SetOutput(logwriter)
	// }

	// // Now from anywhere else in your program, you can use this:
	// log.Print("Starting Bureau of Meteorology Fetcher")
}

// Trace.Println("I have something standard to say")
// Info.Println("Special Information")
// Warning.Println("There is something you need to know about")
// Error.Println("Something has failed")
func initLogHandlers(
	traceHandle io.Writer,
	infoHandle io.Writer,
	warningHandle io.Writer,
	errorHandle io.Writer) {

	Trace = log.New(traceHandle,
		"TRACE: ",
		log.Ldate|log.Ltime|log.Lshortfile)

	Info = log.New(infoHandle,
		"INFO: ",
		log.Ldate|log.Ltime|log.Lshortfile)

	Warning = log.New(warningHandle,
		"WARNING: ",
		log.Ldate|log.Ltime|log.Lshortfile)

	Error = log.New(errorHandle,
		"ERROR: ",
		log.Ldate|log.Ltime|log.Lshortfile)
}

var ioChan chan IO

var mesa device
var ioJson ioStatus

// var loggerConnections map[*websocket.Conn]bool

func main() {
	flag.Parse()
	mesa = newDevice()
	ioChan = make(chan IO)

	initLogHandlers(ioutil.Discard, os.Stdout, os.Stdout, os.Stderr)

	go func() {
		sendNow := make(chan int)
		var count int
		for {
			var lastMsg ioStatus
			lastMsg = ioJson
			select {
			case <-sendNow:
				b, err := json.Marshal(lastMsg)
				if err != nil {
					fmt.Println(err)
				}

				// jm := JsonEncode(ioJson)
				sendIO(b)
				count = 0

			case io := <-ioChan:
				ioJson.Lock()
				ioJson.CarouselPosition = CarouselPosition(io)
				ioJson.Intputs.OrientPinRetractedLS = OrientPinRetractedLS.isOn(io)
				ioJson.Intputs.OrientPinInsertedLS = OrientPinInsertedLS.isOn(io)
				ioJson.Intputs.RotateForwardRapidLS = RotateForwardRapidLS.isOn(io)
				ioJson.Intputs.RotateReverseRapidLS = RotateReverseRapidLS.isOn(io)
				ioJson.Intputs.AdvancedLS = AdvancedLS.isOn(io)
				ioJson.Intputs.RetractedLS = RetractedLS.isOn(io)
				ioJson.Intputs.AdvanceSlowLS = AdvanceSlowLS.isOn(io)
				ioJson.Intputs.RetractSlowLS = RetractSlowLS.isOn(io)
				ioJson.Intputs.InLS = InLS.isOn(io)
				ioJson.Intputs.OutLS = OutLS.isOn(io)
				ioJson.Intputs.SpindleToolUnClampLS = SpindleToolUnClampLS.isOn(io)
				ioJson.Intputs.SpindleToolClampLS = SpindleToolClampLS.isOn(io)
				ioJson.Intputs.CarouselPositionSprocket = CarouselPositionSprocket.isOn(io)
				ioJson.Intputs.UnClampedLS = UnClampedLS.isOn(io)
				ioJson.Intputs.CarouselPositionBit0 = CarouselPositionBit0.isOn(io)
				ioJson.Intputs.CarouselPositionBit1 = CarouselPositionBit1.isOn(io)
				ioJson.Intputs.CarouselPositionBit2 = CarouselPositionBit2.isOn(io)
				ioJson.Intputs.CarouselPositionBit3 = CarouselPositionBit3.isOn(io)
				ioJson.Intputs.CarouselPositionBit4 = CarouselPositionBit4.isOn(io)
				ioJson.Intputs.CarouselPositionBit5 = CarouselPositionBit5.isOn(io)
				ioJson.Intputs.SpindleGearLowLS = SpindleGearLowLS.isOn(io)
				ioJson.Intputs.SpindleGearHighLS = SpindleGearHighLS.isOn(io)
				ioJson.Outputs.OrientPin = OrientPin.isOn(io)
				ioJson.Outputs.CarouselRotateForwardRapid = CarouselRotateForwardRapid.isOn(io)
				ioJson.Outputs.CarouselRotateReverseRapid = CarouselRotateReverseRapid.isOn(io)
				ioJson.Outputs.CarouselRotateForwardSlow = CarouselRotateForwardSlow.isOn(io)
				ioJson.Outputs.CarouselRotateReverseSlow = CarouselRotateReverseSlow.isOn(io)
				ioJson.Outputs.AdvanceRapid = AdvanceRapid.isOn(io)
				ioJson.Outputs.RetractRapid = RetractRapid.isOn(io)
				ioJson.Outputs.AdvanceSlow = AdvanceSlow.isOn(io)
				ioJson.Outputs.RetractSlow = RetractSlow.isOn(io)
				ioJson.Outputs.ZAxisClamp = ZAxisClamp.isOn(io)
				ioJson.Outputs.In = In.isOn(io)
				ioJson.Outputs.Out = Out.isOn(io)
				ioJson.Outputs.UnclampPin = UnclampPin.isOn(io)
				ioJson.Outputs.SpindleToolUnClamp = SpindleToolUnClamp.isOn(io)

				ioJson.Unlock()
				if ioJson != lastMsg {
					b, err := json.Marshal(ioJson)
					if err != nil {
						fmt.Println(err)
					}

					// jm := JsonEncode(ioJson)
					sendIO(b)
				}

				if count > 10 {
					go func() { sendNow <- count }()
				}
				count++

			}
		}
	}()

	go mesa.run()
	loggerConnections = make(map[*websocket.Conn]bool)
	http.HandleFunc("/ws/log", LOGWebSocketsHandler)
	ioConnections = make(map[*websocket.Conn]bool)
	http.HandleFunc("/ws/io", IOWebSocketsHandler)
	manConnections = make(map[*websocket.Conn]bool)
	http.HandleFunc("/ws/man", ManualWebSocketsHandler)
	http.HandleFunc("/io", handleIO)
	http.Handle("/", http.Handler(http.FileServer(http.Dir("www/"))))

	// http.HandleFunc("/", handleRoot)
	log.Fatal(http.ListenAndServe(*listen, nil))

}

type stateFn func(d *device) stateFn

// runs the state Machine
func (d *device) run() {
	for state := Manual; state != nil; {
		state = state(d)
	}
}
