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
	"math/rand"
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
	"0000 - SYS": {
		{
			// just make sure it doesn't explode
			0x0000,
			nil,
			nil,
		},
	},
	"00E0 - CLS": {
		{
			// just make sure it doesn't explode
			0x00E0,
			nil,
			nil,
		},
	},
	"00EE - RET": {
		{
			0x00EE,
			func(t *testing.T, cpu *CPU) {
				cpu.SP = 2
				cpu.Stack[cpu.SP] = 0xFF
			},
			func(t *testing.T, cpu *CPU) {
				assertEquals(t, "PC", cpu.PC, 0x0FF)
				assertEquals(t, "SP", cpu.SP, 0x1)
			},
		},
	},
	"1nnn - JP ADDR": {
		{
			0x10FF,
			nil,
			func(t *testing.T, cpu *CPU) {
				assertEquals(t, "PC", cpu.PC, 0x0FF)
			},
		},
	},
	"2nnn - CALL ADDR": {
		{
			0x2100,
			nil,
			func(t *testing.T, cpu *CPU) {
				assertEquals(t, "SP", cpu.SP, 0x1)
				assertEquals(t, "Stack[1]", cpu.Stack[1], 0x202)
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
	"9xy0 - SNE Vx, Vy": {
		{
			0x9120,
			func(t *testing.T, cpu *CPU) {
				cpu.V[1] = 0x01
				cpu.V[2] = 0x01
			},
			func(t *testing.T, cpu *CPU) {
				assertEquals(t, "PC", cpu.PC, 0x202)
			},
		},
		{
			0x9120,
			func(t *testing.T, cpu *CPU) {
				cpu.V[1] = 0x01
				cpu.V[2] = 0x02
			},
			func(t *testing.T, cpu *CPU) {
				assertEquals(t, "PC", cpu.PC, 0x204)
			},
		},
	},
	"0xAnnn - LD I, addr": {
		{
			0xAFFF,
			nil,
			func(t *testing.T, cpu *CPU) {
				assertEquals(t, "I", cpu.I, 0xFFF)
			},
		},
	},
	"0xBnnn - JP V0, addr": {
		{
			0xBFF0,
			func(t *testing.T, cpu *CPU) {
				cpu.V[0] = 0xF
			},
			func(t *testing.T, cpu *CPU) {
				assertEquals(t, "PC", cpu.PC, 0xFFF)
			},
		},
	},
	"0xCxkk - RND Vx, byte": {
		{
			0xC1FF,
			func(t *testing.T, cpu *CPU) {
				rand.Seed(1) // fix a seed for our RNG call
			},
			func(t *testing.T, cpu *CPU) {
				assertEquals(t, "V1", cpu.V[1], 0x056)
			},
		},
		{
			0xC100,
			func(t *testing.T, cpu *CPU) {
				rand.Seed(1) // fix a seed for our RNG call
			},
			func(t *testing.T, cpu *CPU) {
				assertEquals(t, "V1", cpu.V[1], 0x000)
			},
		},
	},
	"0xEx9E - SKP Vx": {
		{
			0xE19E,
			func(t *testing.T, cpu *CPU) {
				cpu.V[1] = 0x1
			},
			func(t *testing.T, cpu *CPU) {
				assertEquals(t, "PC", cpu.PC, 0x202)
			},
		},
		{
			0xE19E,
			func(t *testing.T, cpu *CPU) {
				cpu.V[1] = 0x1
				cpu.Keypad[0x1] = true
			},
			func(t *testing.T, cpu *CPU) {
				assertEquals(t, "PC", cpu.PC, 0x204)
			},
		},
	},
	"0xExA1 - SKNP Vx": {
		{
			0xE1A1,
			func(t *testing.T, cpu *CPU) {
				cpu.V[1] = 0x1
			},
			func(t *testing.T, cpu *CPU) {
				assertEquals(t, "PC", cpu.PC, 0x204)
			},
		},
		{
			0xE1A1,
			func(t *testing.T, cpu *CPU) {
				cpu.V[1] = 0x1
				cpu.Keypad[0x1] = true
			},
			func(t *testing.T, cpu *CPU) {
				assertEquals(t, "PC", cpu.PC, 0x202)
			},
		},
	},
	"0xFx07 - LD Vx, DT": {
		{
			0xF107,
			func(t *testing.T, cpu *CPU) {
				cpu.DT = 0xFF
			},
			func(t *testing.T, cpu *CPU) {
				assertEquals(t, "V1", cpu.V[1], 0xFF)
			},
		},
	},
	"0xFx15 - LD DT, Vx": {
		{
			0xF115,
			func(t *testing.T, cpu *CPU) {
				cpu.V[1] = 0xFF
			},
			func(t *testing.T, cpu *CPU) {
				assertEquals(t, "DT", cpu.DT, 0xFF)
			},
		},
	},
	"0xFx18 - LD ST, Vx": {
		{
			0xF118,
			func(t *testing.T, cpu *CPU) {
				cpu.V[1] = 0xFF
			},
			func(t *testing.T, cpu *CPU) {
				assertEquals(t, "ST", cpu.ST, 0xFF)
			},
		},
	},
	"0xFx1E - ADD I, Vx": {
		{
			0xF11E,
			func(t *testing.T, cpu *CPU) {
				cpu.I = 0xF0
				cpu.V[1] = 0x0F
			},
			func(t *testing.T, cpu *CPU) {
				assertEquals(t, "I", cpu.I, 0xFF)
			},
		},
	},
	"Fx29 - LD F, Vx": {
		{
			0xF129,
			func(t *testing.T, cpu *CPU) {
				cpu.V[1] = 0x01
			},
			func(t *testing.T, cpu *CPU) {
				assertEquals(t, "I", cpu.I, 0x05)
			},
		},
	},
	"Fx33 - LD B, Vx": {
		{
			0xF133,
			func(t *testing.T, cpu *CPU) {
				cpu.I = 0xFF
				cpu.V[1] = 0xFF
			},
			func(t *testing.T, cpu *CPU) {
				assertEquals(t, "M[I+0]", cpu.Memory[cpu.I+0], 2)
				assertEquals(t, "M[I+1]", cpu.Memory[cpu.I+1], 5)
				assertEquals(t, "M[I+2]", cpu.Memory[cpu.I+2], 5)
			},
		},
	},
	"Fx55 - LD [I], Vx": {
		{
			0xF255,
			func(t *testing.T, cpu *CPU) {
				cpu.I = 0x16
				cpu.V[0] = 0xB
				cpu.V[1] = 0xA
				cpu.V[2] = 0xD
			},
			func(t *testing.T, cpu *CPU) {
				assertEquals(t, "M[I+0]", cpu.Memory[cpu.I+0], cpu.V[0])
				assertEquals(t, "M[I+1]", cpu.Memory[cpu.I+1], cpu.V[1])
				assertEquals(t, "M[I+2]", cpu.Memory[cpu.I+2], cpu.V[2])
			},
		},
	},
	"Fx65 - LD [I], Vx": {
		{
			0xF265,
			func(t *testing.T, cpu *CPU) {
				cpu.I = 0x16
				cpu.Memory[cpu.I+0] = 0xB
				cpu.Memory[cpu.I+1] = 0xA
				cpu.Memory[cpu.I+2] = 0xD
			},
			func(t *testing.T, cpu *CPU) {
				assertEquals(t, "V[0]", cpu.V[0], cpu.Memory[cpu.I+0])
				assertEquals(t, "V[1]", cpu.V[1], cpu.Memory[cpu.I+1])
				assertEquals(t, "V[2]", cpu.V[2], cpu.Memory[cpu.I+2])
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
				// add some more detail to failing tests
				if t.Failed() {
					t.Logf("Instruction: %s", label)
					t.Logf("Opcode: 0x%04X", test.Opcode)
					t.FailNow()
				}
			}
			// spin off a sub-test
			t.Run(fmt.Sprintf("%s/%d", label, index), executeTest)
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
