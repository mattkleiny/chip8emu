// This package implements a simple Chip 8 interpreter for the Chip 8.
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
	Keypad [16]byte   // 16-key hexadecimal keypad.
	Pixels Bitmap     // The pixel bitmap representing the display output.
}

// Represents a bitmap of pixels as used in our Chip 8 implementation.
// 64 * 32 pixels (2048 total pixels). The origin (0, 0) is in the top left.
type Bitmap [Width * Height]byte

// Empties the bitmap's content.
func (bitmap *Bitmap) clear() {
	for y := 0; y < Height-1; y++ {
		for x := 0; x < Width-1; x++ {
			bitmap[x+y*Width] = 0
		}
	}
}

// Writes a sprite at the given (x, y) coordinates
// Returns a flag indicating if an existing pixel was overwritten.
func (bitmap *Bitmap) writeSprite(sprite []byte, x, y byte) (collision bool) {
	// clamps the given unsigned value below the given maximum
	clamp := func(value, max byte) byte {
		if value > max {
			return value - max
		}
		return value
	}

	// walk over the sprite
	for j := 0; j < len(sprite); j++ {
		for i := 0; i < 8; i++ {
			xpos := clamp(x+byte(i), Width)
			ypos := clamp(y+byte(j), Height)

			pixel := &bitmap[xpos+ypos*Width]

			panic("TODO")
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

// Initializes a new CPU.
func NewCPU() *CPU {
	cpu := new(CPU)
	// programs expected to start at 0x200
	cpu.PC = 0x200
	// load the font-set
	for i := 0; i < len(fontSet); i++ {
		cpu.Memory[i] = fontSet[i]
	}
	return cpu
}

// Loads a program into the CPU.
func (cpu *CPU) LoadProgram(program []byte) {
	for i := 0; i < len(program); i++ {
		cpu.Memory[i+0x200] = program[i] // programs are expected to start at 0x200
	}
}

// Advances the CPU a single cycle.
func (cpu *CPU) NextCycle() {
	// fetch the next instruction based on the program counter
	opcode := uint16(cpu.Memory[cpu.PC])<<8 | uint16(cpu.Memory[cpu.PC+1])

	// move to the next instruction
	cpu.PC += 2

	// extract common operands from the opcode
	x := byte((opcode & 0x0F00) >> 8)
	y := byte((opcode & 0x00F0) >> 4)
	kk := byte(opcode)
	n := opcode & 0x000F
	nnn := opcode & 0xFFF

	// pointers for commonly accessed registers
	Vx := &cpu.V[x]
	Vy := &cpu.V[y]
	VF := &cpu.V[0xF]

	// advance timers by a single cycle
	advanceTimers := func() {
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
	defer advanceTimers()

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
		cpu.Stack[cpu.SP] = cpu.PC
		cpu.SP += 1
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
				*VF = 1 // carry
			} else {
				*VF = 0
			}
			*Vx += *Vy

		case 0x0005: // SUB Vx, Vy
			if *Vy > (0xFF - *Vx) {
				*VF = 1 // carry
			} else {
				*VF = 0
			}
			*Vx -= *Vy

		case 0x0006: // SHR Vx
			panic("TODO")

		case 0x0007: // SUBN Vx, Vy
			panic("TODO")

		case 0x000E: // SHL Vx
			panic("TODO")
		}

	case 0x9000: // SNE Vx, Vy
		panic("TODO")

	case 0xA000: // LD I, addr
		cpu.I = nnn

	case 0xB000: // JP V0, addr
		cpu.PC = nnn + uint16(cpu.V[0])

	case 0xC000: // RND Vx, byte
		*Vx = kk + randomByte()

	case 0xD000: // DRW Vx, Vy, nibble
		// sample the sprite and render it at the (X, Y) coordinates
		sprite := cpu.Memory[cpu.I:cpu.I+n]
		cpu.Pixels.writeSprite(sprite, x, y)

	case 0x0033: // LD B, Vx
		cpu.Memory[cpu.I] = cpu.V[*Vx] / 100
		cpu.Memory[cpu.I+1] = (cpu.V[*Vx] / 10) % 10
		cpu.Memory[cpu.I+2] = (cpu.V[*Vx] % 100) % 10

	default:
		println("Unknown opcode: ", opcode)
	}
}

// generates a random byte (0 to 255, inclusive).
func randomByte() byte {
	source := rand.NewSource(time.Now().UnixNano())
	rng := rand.New(source)
	value := rng.Intn(255)

	return byte(value)
}
