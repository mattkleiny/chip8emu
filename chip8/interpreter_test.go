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

// Tests for all potential opcodes in the system.
// This test case set-up is based on ejholmes implementation here:
// https://github.com/ejholmes/chip8/blob/master/chip8_test.go.
var OpcodeTests = map[string][]OpcodeTest{
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
				assertEquals(t, "PC", cpu.PC, 0x204)
			},
		},
	},
}

// Asserts that all of the opcodes execute as expected in the CPU
func TestOpcodes(t *testing.T) {
	// for each opcode scenario
	for label, tests := range OpcodeTests {
		// execute each individual test
		for index, test := range tests {
			executeTest := func(t *testing.T) {
				cpu := NewCPU() // allocate a new CPU
				// before handling
				if test.Before != nil {
					test.Before(t, cpu)
				}
				// decode and execute
				cpu.decodeAndExecute(test.Opcode)
				// after handling
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
