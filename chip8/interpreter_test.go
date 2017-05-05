// Copyright © 2017 Matthew Kleinschafer
// 
// Permission is hereby granted, free of charge, to any person
// obtaining a copy of this software and associated documentation
// files (the “Software”), to deal in the Software without
// restriction, including without limitation the rights to use,
// copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the
// Software is furnished to do so, subject to the following
// conditions:
// 
// The above copyright notice and this permission notice shall be
// included in all copies or substantial portions of the Software.
// 
// THE SOFTWARE IS PROVIDED “AS IS”, WITHOUT WARRANTY OF ANY KIND,
// EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES
// OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND
// NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT
// HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY,
// WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING
// FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
// OTHER DEALINGS IN THE SOFTWARE.

package chip8

import (
	"fmt"
	"testing"
)

// Encapsulates a test case for a single opcode in the CPU
type OpcodeTest struct {
	Opcode uint16
	Before func(t *testing.T, cpu *CPU)
	After  func(t *testing.T, cpu *CPU)
}

// A scenario of tests for a particular opcode type.
type OpcodeScenario []OpcodeTest

// Tests for all potential opcodes in the system.
// This test case set-up is based on ejholmes implementation here:
// https://github.com/ejholmes/chip8/blob/master/chip8_test.go.
var OpcodeScenarios = map[string]OpcodeScenario{
	"2nnn - CALL ADDR": {
		{
			0x2100,
			nil,
			func(t *testing.T, cpu *CPU) {
				assertEquals(t, "Stack[0]", cpu.Stack[0], 0x200)
				assertEquals(t, "SP", cpu.SP, 0x1)
				assertEquals(t, "PC", cpu.PC, 0x100)
			},
		},
	},
	"3xkk - SE Vx, byte": {
		{
			0x3123,
			nil,
			func(t *testing.T, cpu *CPU) {
				assertEquals(t, "PC", cpu.PC, 0x202)
			},
		},
		{
			0x3103,
			func(t *testing.T, cpu *CPU) {
				cpu.V[1] = 0x03
			},
			func(t *testing.T, cpu *CPU) {
				assertEquals(t, "PC", cpu.PC, 0x204)
			},
		},
	},
	"4xkk - SNE Vx, byte": {
		{
			0x4123,
			nil,
			func(t *testing.T, cpu *CPU) {
				assertEquals(t, "PC", cpu.PC, 0x204)
			},
		},
		{
			0x4103,
			func(t *testing.T, cpu *CPU) {
				cpu.V[1] = 0x03
			},
			func(t *testing.T, cpu *CPU) {
				assertEquals(t, "PC", cpu.PC, 0x202)
			},
		},
	},
	"5xy0 - SE Vx, Vy": {
		{
			0x5120,
			func(t *testing.T, cpu *CPU) {
				cpu.V[1] = 0x40
			},
			func(t *testing.T, cpu *CPU) {
				assertEquals(t, "PC", cpu.PC, 0x202)
			},
		},
		{
			0x5120,
			func(t *testing.T, cpu *CPU) {
				cpu.V[1] = 0x03
				cpu.V[2] = 0x03
			},
			func(t *testing.T, cpu *CPU) {
				assertEquals(t, "PC", cpu.PC, 0x204)
			},
		},
	},
	"6xkk - LD Vx, byte": {
		{
			0x6123,
			nil,
			func(t *testing.T, cpu *CPU) {
				assertEquals(t, "V1", cpu.V[1], 0x23)
			},
		},
	},
	"7xkk - ADD Vx, byte": {
		{
			0x7101,
			func(t *testing.T, cpu *CPU) {
				cpu.V[1] = 01
			},
			func(t *testing.T, cpu *CPU) {
				assertEquals(t, "V1", cpu.V[1], 0x02)
			},
		},
	},
	"8xy0 - LD Vx, Vy": {
		{
			0x8120,
			func(t *testing.T, cpu *CPU) {
				cpu.V[1] = 0x00
				cpu.V[2] = 0xFF
			},
			func(t *testing.T, cpu *CPU) {
				assertEquals(t, "V1", cpu.V[1], 0xFF)
			},
		},
	},
	"8xy1 - OR Vx, Vy": {
		{
			0x8121,
			func(t *testing.T, cpu *CPU) {
				cpu.V[1] = 0x00
				cpu.V[2] = 0xF0
			},
			func(t *testing.T, cpu *CPU) {
				assertEquals(t, "V1", cpu.V[1], 0xF0)
			},
		},
	},
	"8xy2 - AND Vx, Vy": {
		{
			0x8122,
			func(t *testing.T, cpu *CPU) {
				cpu.V[1] = 0x0F
				cpu.V[2] = 0xFF
			},
			func(t *testing.T, cpu *CPU) {
				assertEquals(t, "V1", cpu.V[1], 0x0F)
			},
		},
	},
	"8xy3 - XOR Vx, Vy": {
		{
			0x8123,
			func(t *testing.T, cpu *CPU) {
				cpu.V[1] = 0x0F
				cpu.V[2] = 0xF0
			},
			func(t *testing.T, cpu *CPU) {
				assertEquals(t, "V1", cpu.V[1], 0xFF)
			},
		},
	},
	// TODO: check VF flag in these ops
	"8xy4 - ADD Vx, Vy": {
		{
			0x8124,
			func(t *testing.T, cpu *CPU) {
				cpu.V[1] = 0x01
				cpu.V[2] = 0x02
			},
			func(t *testing.T, cpu *CPU) {
				assertEquals(t, "V1", cpu.V[1], 0x03)
			},
		},
	},
	"8xy5 - SUB Vx, Vy": {
		{
			0x8125,
			func(t *testing.T, cpu *CPU) {
				cpu.V[1] = 0x02
				cpu.V[2] = 0x01
			},
			func(t *testing.T, cpu *CPU) {
				assertEquals(t, "V1", cpu.V[1], 0x01)
			},
		},
	},
	"8xy6 - SHR Vx {, Vy}": {
		{
			0x8106,
			func(t *testing.T, cpu *CPU) {
				cpu.V[1] = 0x04
			},
			func(t *testing.T, cpu *CPU) {
				assertEquals(t, "V1", cpu.V[1], 0x02)
				assertEquals(t, "VF", cpu.V[0xF], 0)
			},
		},
	},
	"8xy7 - SUBN Vx, Vy": {
		{
			0x8127,
			func(t *testing.T, cpu *CPU) {
				cpu.V[1] = 0x04
				cpu.V[2] = 0x04
			},
			func(t *testing.T, cpu *CPU) {
				assertEquals(t, "V1", cpu.V[1], 0)
				assertEquals(t, "VF", cpu.V[0xF], 0)
			},
		},
	},
	"8xyE - SHL Vx {, Vy}": {
		{
			0x810E,
			func(t *testing.T, cpu *CPU) {
				cpu.V[1] = 0x02
			},
			func(t *testing.T, cpu *CPU) {
				assertEquals(t, "V1", cpu.V[1], 0x04)
				assertEquals(t, "VF", cpu.V[0xF], 0)
			},
		},
	},
}

// Asserts that all of the opcodes execute as expected in the CPU
func TestOpcodes(t *testing.T) {
	// for each opcode scenario
	for label, scenario := range OpcodeScenarios {
		// execute each individual test in the scenario
		for index, test := range scenario {
			executeTest := func(t *testing.T) {
				cpu := NewCPU() // allocate a new CPU for each test scenario
				if test.Before != nil {
					test.Before(t, cpu)
				}
				cpu.decodeAndExecute(test.Opcode)
				if test.After != nil {
					test.After(t, cpu)
				}
			}
			// spin off a sub-test
			t.Run(fmt.Sprintf("%s/%d", label, index), executeTest)
		}
	}
}

// Ensure we're able to reset the display bitmap completely.
func TestClearBitmap(t *testing.T) {
	bitmap := new(Bitmap)

	// fill with junk
	for x := 0; x < Width-1; x++ {
		for y := 0; y < Height-1; y++ {
			bitmap.setPixel(x, y, randomByte())
		}
	}

	// clear the thing, assert it's empty
	bitmap.clear()
	for x := 0; x < Width-1; x++ {
		for y := 0; y < Height-1; y++ {
			if bitmap.getPixel(x, y) != 0 {
				t.Error("The bitmap was not cleared successfully")
			}
		}
	}
}

// Ensure we're able to write a sprite to the display bitmap
func TestWriteSprite(t *testing.T) {
	const size = 3    // square size of the sprite
	const offsetX = 3 // x offset for resultant sprite
	const offsetY = 6 // y offset for resultant sprite

	sprite := []byte{
		0x1, 0x2, 0x3,
		0x4, 0x5, 0x6,
		0x7, 0x8, 0x9,
	}

	// write the sprite into the bitmap
	bitmap := new(Bitmap)
	bitmap.writeSprite(sprite, offsetX, offsetY)

	// check each of the resultant pixels
	for x := offsetX; x < size+offsetX; x++ {
		for y := offsetY; y < size+offsetY; y++ {
			if bitmap.getPixel(x, y) != sprite[(x-offsetX)+(y-offsetY)*size] {
				t.Errorf("The pixel byte at (%d, %d) does not match the expected", x, y)
			}
		}
	}
}

// Checks the the given value against the expected
func assertEquals(t *testing.T, subject string, actual, expected interface{}) {
	// Attempts to convert a value to a uint16
	asuint16 := func(value interface{}) uint16 {
		switch value := value.(type) {
		case byte:
			return uint16(value)
		case uint16:
			return value
		case int:
			return uint16(value)
		case uint32:
			return uint16(value)
		}
		return 0
	}

	a := asuint16(actual)
	e := asuint16(expected)

	if a != e {
		t.Errorf("%s was 0x%04X; expected 0x%04X", subject, a, e)
	}
}
