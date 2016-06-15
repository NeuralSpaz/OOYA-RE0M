package main

import (
	"errors"
	"fmt"
	"math"
)

////////////////////////////////////////////////////////////////////////////////////////
// Util Functions

// Determines Which way to rotate
func getDirection(startPos, endPos uint8) (string, error) {
	numberOfPositions := 20.0
	if startPos == 0 || startPos > uint8(numberOfPositions) || startPos < 1 {
		return "No Rotation", errors.New("Error: ATC Direction Start Position invalid")
	}
	if endPos == 0 || endPos > uint8(numberOfPositions) || endPos < 1 {
		return "No Rotation", errors.New("Error: ATC Direction End Position invalid")
	}
	if startPos == endPos {
		return "No Rotation", nil
	}
	distance := math.Abs(float64(startPos - endPos))
	lessThanHalf := distance < numberOfPositions/2

	if lessThanHalf && (endPos < startPos) || !lessThanHalf && (endPos > startPos) {
		return "Reverse", nil
	} else {
		return "Forward", nil
	}
}

var ErrNotBCD = errors.New("Byte is Not BCD Encoded")

// uint8 to Packed BCD 8-4-2-1 One digit per nibble
func Uint8toBCD(u uint8) byte {
	lsn := u % 10
	u /= 10
	msn := u % 10
	return ((msn & 0xf) << 4) | (lsn & 0xf)
}

// Packed BCD 8-4-2-1 One digit per nibble to uint8
// Error if not a BCD digits
func BCDtoUint8(bcd byte) (uint8, error) {
	digits := uint8((bcd>>4&0xf)*10 + (bcd & 0xf))
	// Confirm input is BCD encoded as expected
	check := Uint8toBCD(digits)

	if bcd != check|bcd {
		return digits, ErrNotBCD
	}
	return digits, nil
}

func (p input) isOn(io IO) bool {
	return (io.inputs & input(p)) > 0
}

func (p output) isOn(io IO) bool {
	return (io.outputs & output(p)) > 0
}

func PositionIsValid(io IO) bool {
	positionByte := CarouselPositionBit0 & io.inputs >> 16
	positionByte += CarouselPositionBit1 & io.inputs >> 16
	positionByte += CarouselPositionBit2 & io.inputs >> 16
	positionByte += CarouselPositionBit3 & io.inputs >> 16
	positionByte += CarouselPositionBit4 & io.inputs >> 16
	positionByte += CarouselPositionBit5 & io.inputs >> 16
	pos, err := BCDtoUint8(uint8(positionByte))
	fmt.Printf("%8.8b\n", positionByte)
	// fmt.Println("CarouselPosition :", pos)
	if err != nil {
		fmt.Println(pos)
		return false
	}
	if pos < 1 || pos > 20 {
		return false
	}

	return true
}

func CarouselPosition(io IO) int {
	positionByte := CarouselPositionBit0 & io.inputs >> 16
	positionByte += CarouselPositionBit1 & io.inputs >> 16
	positionByte += CarouselPositionBit2 & io.inputs >> 16
	positionByte += CarouselPositionBit3 & io.inputs >> 16
	positionByte += CarouselPositionBit4 & io.inputs >> 16
	positionByte += CarouselPositionBit5 & io.inputs >> 16
	pos, _ := BCDtoUint8(uint8(positionByte))
	return int(pos)
}
