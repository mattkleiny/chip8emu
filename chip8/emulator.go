// This package implements a simple emulator for the Chip 8 system/interpreter
// See see http://devernay.free.fr/hacks/chip8/C8TECH10.HTM for more detail.
package chip8

import (
	"math/rand"
	"time"
)

const (
	Width  = 64 // Display width, in pixels
	Height = 32 // Display height, in pixels
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
	Memory   [4096]byte           // The Chip 8 has fixed 4K memory in total
	V        [16]byte             // 15 8-bit general purpose registers (V0, V1 through to VE). The 16th is the carry flag
	Stack    [16]uint16           // The currently executing instruction; references the program counter
	I        uint16               // An index register
	PC       uint16               // A program counter PC which can have values from 0x000 to 0xFFF
	SP       byte                 // The stack pointer
	DT, ST   byte                 // When above zero, they count down to zero. Counting occurs at 60hz
	Keypad   [16]byte             // 16-key hexadecimal keypad
	Pixels   [Width * Height]byte // 64 * 32 pixels (2048 total pixels). The origin (0, 0) is in the top left.
	DrawFlag bool                 // True whether the display has been updated this cycle
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

// Initializes a new CPU
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

// Loads a program into the CPU
func (cpu *CPU) LoadProgram(program []byte) {
	for i := 0; i < len(program); i++ {
		cpu.Memory[i+0x200] = program[i] // programs are expected to start at 0x200
	}
}

// Advances the CPU a single cycle
func (cpu *CPU) NextCycle() {
	// fetch the next instruction based on the program counter
	opcode := uint16(cpu.Memory[cpu.PC])<<8 | uint16(cpu.Memory[cpu.PC+1])

	// move to the next instruction
	cpu.PC += 2

	// extract common operands from the opcode
	x := (opcode & 0x0F00) >> 8
	y := (opcode & 0x00F0) >> 4
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

	// generates a random byte
	randomByte := func() byte {
		source := rand.NewSource(time.Now().UnixNano())
		rng := rand.New(source)
		value := rng.Intn(255)

		return byte(value)
	}

	// clamps the given unsigned value below the given maximum
	clamp := func(value, max uint) uint {
		if value > max {
			return value - max
		}
		return value
	}

	// decode and execute the opcode
	switch opcode & 0xF000 {
	case 0x0000:
		switch opcode {
		case 0x00E0: // CLS
			for y := 0; y < Height-1; y++ {
				for x := 0; x < Width-1; x++ {
					cpu.Pixels[x+y*Width] = 0
				}
			}

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
		// sample the sprite
		sprite := cpu.Memory[cpu.I:cpu.I+n]

		for y := 0; y < len(sprite); y++ {
			row := sprite[y]

			for x := 0; x < 8; x++ {
				xpos := clamp(uint(x)+uint(*Vx), Width)
				ypos := clamp(uint(y)+uint(*Vy), Height)

				// check to see if the pixel is already illuminated
				i := byte(0x80 >> byte(x))

				// adjust the pixel with an XOR
				pixel := &cpu.Pixels[xpos+ypos*Width]

				if *pixel != 0 {
					*VF = 1 // carry if overwritten
				}

				if row&i == i {
					*pixel = *pixel ^ 1
				} else {
					*pixel = *pixel ^ 0
				}
			}
		}

	case 0x0033: // LD B, Vx
		cpu.Memory[cpu.I] = cpu.V[*Vx] / 100
		cpu.Memory[cpu.I+1] = (cpu.V[*Vx] / 10) % 10
		cpu.Memory[cpu.I+2] = (cpu.V[*Vx] % 100) % 10

	default:
		println("Unknown opcode: ", opcode)
	}
}
