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

// This package implements a Chip 8 interpreter.
// See see http://devernay.free.fr/hacks/chip8/C8TECH10.HTM for more detail.
package chip8

import (
	"math/rand"
	"time"
)

const (
	Width  = 64 // Display width, in pixels.
	Height = 32 // Display height, in pixels.
)

// The central processing unit of the chip 8 system
// Memory is laid-out in the following structure:
// +---------------+= 0xFFF (4095) End of Chip-8 RAM
// |               |
// |               |
// |               |
// |               |
// |               |
// | 0x200 to 0xFFF|
// |     Chip-8    |
// | Program / Data|
// |     Space     |
// |               |
// |               |
// |               |
// +- - - - - - - -+= 0x600 (1536) Start of ETI 660 Chip-8 programs
// |               |
// |               |
// |               |
// +---------------+= 0x200 (512) Start of most Chip-8 programs
// | 0x000 to 0x1FF|
// | Reserved for  |
// |  interpreter  |
// +---------------+= 0x000 (0) Start of Chip-8 RAM
type CPU struct {
	Memory [4096]byte // The Chip 8 has fixed 4K memory in total.
	V      [16]byte   // 15 8-bit general purpose registers (V0, V1 through to VE). The 16th is the carry flag.
	I      uint16     // A 16-bit register.
	PC     uint16     // A program counter PC (which can have values from 0x000 to 0xFFF).
	SP     byte       // The stack pointer.
	Stack  [16]uint16 // The stack of branching instructions; references the program counter.
	DT, ST byte       // Delay/Sound timers. When above zero, they count down to zero. Counting occurs at 60hz.
	Keypad [16]bool   // 16-key hexadecimal keypad. Boolean flag indicates if key is pressed or not.
	Pixels Bitmap     // The pixel bitmap representing the display output.
}

// Represents a bitmap of pixels as used in our Chip 8 implementation.
// 64 * 32 pixels (2048 total pixels). The origin (0, 0) is in the top left.
type Bitmap [Width * Height]byte

// Retrieves the pixel value at the given (x, y) coordinates.
func (bitmap *Bitmap) getPixel(x, y int) byte {
	return bitmap[x+y*Width]
}

// Sets the pixel value at the given (x, y) coordinates.
func (bitmap *Bitmap) setPixel(x, y int, value byte) {
	bitmap[x+y*Width] = value
}

// Empties the bitmap's content.
func (bitmap *Bitmap) clear() {
	for y := 0; y < Height-1; y++ {
		for x := 0; x < Width-1; x++ {
			bitmap[x+y*Width] = 0
		}
	}
}

// Writes a sprite at the given (x, y) coordinates.
// A sprite is a collection of bits representing pixel values over a range.
// Returns a flag indicating if an existing pixel was overwritten.
func (bitmap *Bitmap) writeSprite(sprite []byte, x, y byte) (collision bool) {
	// wraps the given unsigned value around the given maximum
	wrap := func(value, max byte) byte {
		if value > max {
			return value - max
		}
		return value
	}

	// walk over the sprite
	for j := 0; j < len(sprite); j++ {
		row := sprite[j]
		for i := 0; i < 8; i++ { // 8 bits per sprite byte
			// calculate wrapped x and y pixel coordinates
			xpos := wrap(x+byte(i), Width)
			ypos := wrap(y+byte(j), Height)

			pixel := &bitmap[xpos+ypos*Width]

			// see if we're stamping on an existing pixel
			if *pixel == 0x01 {
				collision = true
			}

			// see if we're activating a pixel
			mask := byte(0x80 >> byte(i))
			if (row & mask) == mask {
				*pixel = *pixel ^ 0x01
			} else {
				*pixel = *pixel ^ 0x00
			}
		}
	}

	return
}

// The default font-set for the chip 8 system
//
// Each entry represents a small quad that renders a particular character
// Each character is 4 pixels wide by 5 pixels high
var fontSet = []byte{
	0xF0, 0x90, 0x90, 0x90, 0xF0, // 0
	0x20, 0x60, 0x20, 0x20, 0x70, // 1
	0xF0, 0x10, 0xF0, 0x80, 0xF0, // 2
	0xF0, 0x10, 0xF0, 0x10, 0xF0, // 3
	0x90, 0x90, 0xF0, 0x10, 0x10, // 4
	0xF0, 0x80, 0xF0, 0x10, 0xF0, // 5
	0xF0, 0x80, 0xF0, 0x90, 0xF0, // 6
	0xF0, 0x10, 0x20, 0x40, 0x40, // 7
	0xF0, 0x90, 0xF0, 0x90, 0xF0, // 8
	0xF0, 0x90, 0xF0, 0x10, 0xF0, // 9
	0xF0, 0x90, 0xF0, 0x90, 0x90, // A
	0xE0, 0x90, 0xE0, 0x90, 0xE0, // B
	0xF0, 0x80, 0x80, 0x80, 0xF0, // C
	0xE0, 0x90, 0x90, 0x90, 0xE0, // D
	0xF0, 0x80, 0xF0, 0x80, 0xF0, // E
	0xF0, 0x80, 0xF0, 0x80, 0x80, // F
}

// Programs are expected to start at 0x200.
const startOffset = 0x200

// Initializes a new CPU.
func NewCPU() *CPU {
	cpu := new(CPU)
	// programs expected to start at 0x200
	cpu.PC = startOffset
	// load the font-set
	for i := 0; i < len(fontSet); i++ {
		cpu.Memory[i] = fontSet[i]
	}
	return cpu
}

// Loads a program into the CPU from the given byte slice.
func (cpu *CPU) LoadProgram(program []byte) {
	for i := 0; i < len(program); i++ {
		cpu.Memory[i+startOffset] = program[i]
	}
}

// Runs the CPU at the given frequency, in hertz
func (cpu *CPU) RunAtFrequency(frequency uint) {
	// tick at a fixed interval (roughly)
	clock := time.Tick(time.Second / time.Duration(frequency))
	// run forever, or until an exit signal is detected
	running := true
	for running {
		select {
		case <-clock:
			cpu.NextCycle() // advance the cpu
		}
	}
}

// Advances the CPU a single cycle.
func (cpu *CPU) NextCycle() {
	// fetch the next instruction based on the program counter
	opcode := uint16(cpu.Memory[cpu.PC])<<8 | uint16(cpu.Memory[cpu.PC+1])

	// execute the instruction
	cpu.decodeAndExecute(opcode)

	// advance timers by a single cycle
	if cpu.DT > 0 {
		cpu.DT -= 1
	}
	if cpu.ST > 0 {
		if cpu.ST == 1 {
			println("BEEP")
		}
		cpu.ST -= 1
	}
}

// Decodes and executes the given opcode.
func (cpu *CPU) decodeAndExecute(opcode uint16) {
	// move to the next instruction
	cpu.PC += 2

	// extract common operands from the opcode
	x := byte((opcode & 0x0F00) >> 8)
	y := byte((opcode & 0x00F0) >> 4)
	kk := byte(opcode)
	n := opcode & 0x000F
	nnn := opcode & 0x0FFF

	// pointers for commonly accessed registers
	Vx := &cpu.V[x]
	Vy := &cpu.V[y]
	VF := &cpu.V[0xF]

	// decode and execute the opcode
	switch opcode & 0xF000 {
	case 0x0000:
		switch opcode {
		case 0x00E0: // CLS
			cpu.Pixels.clear()

		case 0x00EE: // RET
			cpu.PC = cpu.Stack[cpu.SP]
			cpu.SP -= 1

		case 0x0000: // SYS addr
			// no-op
			break
		}

	case 0x1000: // JP addr
		cpu.PC = nnn

	case 0x2000: // CALL addr
		cpu.SP += 1
		cpu.Stack[cpu.SP] = cpu.PC - 2
		cpu.PC = nnn

	case 0x3000: // SE Vx, byte
		if *Vx == kk {
			cpu.PC += 2 // skip the next instruction
		}

	case 0x4000: // SNE Vx, byte
		if *Vx != kk {
			cpu.PC += 2 // skip the next instruction
		}

	case 0x5000: // SE Vx, Vy
		if *Vx == *Vy {
			cpu.PC += 2 // skip the next instruction
		}

	case 0x6000: // LD Vx, byte
		*Vx = kk

	case 0x7000: // ADD Vx, byte
		*Vx += kk

	case 0x8000:
		switch opcode & 0x000F {
		case 0x0000: // LD Vx, Vy
			*Vx = *Vy

		case 0x0001: // OR Vx, Vy
			*Vx = *Vy | *Vx

		case 0x0002: // AND Vx, Vy
			*Vx = *Vy & *Vx

		case 0x0003: // XOR Vx, Vy
			*Vx = *Vy ^ *Vx

		case 0x0004: // ADD Vx, Vy
			if *Vy > (0xFF - *Vx) {
				*VF = 1
			} else {
				*VF = 0
			}
			*Vx += *Vy

		case 0x0005: // SUB Vx, Vy
			if *Vy > (0xFF - *Vx) {
				*VF = 1
			} else {
				*VF = 0
			}
			*Vx -= *Vy

		case 0x0006: // SHR Vx {, Vy}
			if (*Vx & 0x01) == 0x01 {
				*VF = 1
			} else {
				*VF = 0
			}
			*Vx /= 2

		case 0x0007: // SUBN Vx, Vy
			if *Vy > *Vx {
				*VF = 1
			}
			*Vx = *Vy - *Vx

		case 0x000E: // SHL Vx {, Vy}
			if (*Vx & 0x80) == 0x80 {
				*VF = 1
			} else {
				*VF = 0
			}
			*Vx *= 2
		}

	case 0x9000: // SNE Vx, Vy
		if *Vx != *Vy {
			cpu.PC += 2
		}

	case 0xA000: // LD I, addr
		cpu.I = nnn

	case 0xB000: // JP V0, addr
		cpu.PC = nnn + uint16(cpu.V[0])

	case 0xC000: // RND Vx, byte
		*Vx = kk & randomByte()

	case 0xD000: // DRW Vx, Vy, nibble
		// sample the sprite and render it at the (X, Y) coordinates
		sprite := cpu.Memory[cpu.I:cpu.I+n]
		if cpu.Pixels.writeSprite(sprite, x, y) {
			*VF = 1
		} else {
			*VF = 0
		}

	case 0xE000: // SKNP Vx
		if !cpu.Keypad[*Vx] {
			cpu.PC += 2
		}

	case 0xF000:
		switch opcode & 0x00FF {
		case 0x0007: // LD Vx, DT
			*Vx = cpu.DT

		case 0x000A: // LD Vx, K
			println("Wait for key input not yet supported")

		case 0x0015: // LD DT, Vx
			cpu.DT = *Vx

		case 0x0018: // LD ST, Vx
			cpu.ST = *Vx

		case 0x001E:
			cpu.I = cpu.I + uint16(*Vx)

		case 0x0033: // LD B, Vx
			cpu.Memory[cpu.I] = *Vx / 100
			cpu.Memory[cpu.I+1] = (*Vx / 10) % 10
			cpu.Memory[cpu.I+2] = (*Vx % 100) % 10

		case 0x0055: // LD [I], Vx
			for i := byte(0); i <= x; i++ {
				cpu.Memory[cpu.I+uint16(i)] = cpu.V[i]
			}

		case 0x0065: // LD Vx, [I]
			for i := byte(0); i <= x; i++ {
				cpu.V[i] = cpu.Memory[cpu.I+uint16(i)]
			}
		}

	default:
		println("Unknown opcode: ", opcode)
	}
}

// generates a random byte (0 to 255, inclusive).
func randomByte() byte {
	return byte(rand.Intn(255))
}
