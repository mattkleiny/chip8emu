// This package implements a simple emulator for the Chip 8 system/interpreter
// See see http://devernay.free.fr/hacks/chip8/C8TECH10.HTM for more detail.
package chip8

type Opcode uint16 // 35 possible opcodes, which are all two bytes long

// The central processing unit of the chip 8 system
type CPU struct {
	Memory [4096]byte // The Chip 8 has fixed 4K memory in total
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

	V     [16]byte   // 15 8-bit general purpose registers (V0, V1 through to VE). The 16th is the carry flag
	Stack [16]uint16 // The currently executing instruction; references the program counter
	I, PC uint16     // An index register I, and a program counter PC which can have values from 0x000 to 0xFFF
	SP    uint16     // The stack pointer

	Keypad   [16]byte      // 16-key hexadecimal keypad
	Pixels   [64 * 32]byte // 64 * 32 pixels (2048 total pixels). The origin (0, 0) is in the top left.
	DrawFlag bool          // True whether the display has been updated this cycle

	DelayTimer, SoundTimer byte // When above zero, they count down to zero. Counting occurs at 60hz
}

// The default font-set for the chip 8 system
var FontSet = []byte{
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

// Initializes the CPU
func NewCpu() *CPU {
	cpu := new(CPU)
	cpu.PC = 0x200 // program counter starts at 0x200

	// load the font-set
	for i := 0; i < len(FontSet); i++ {
		cpu.Memory[i] = FontSet[i]
	}

	return cpu
}

// Loads a program into the system
func (cpu *CPU) LoadProgram(program []byte) {
	for i := 0; i < len(program); i++ {
		cpu.Memory[i+0x200] = program[i] // programs are expected to start at 0x200
	}
}

// Advances the system a single cycle
func (cpu *CPU) NextCycle() {
	// updates the CPU timers
	updateTimers := func() {
		if cpu.DelayTimer > 0 {
			cpu.DelayTimer -= 1
		}
		if cpu.SoundTimer > 0 {
			if cpu.SoundTimer == 1 {
				println("BEEP")
			}
			cpu.SoundTimer -= 1
		}
	}

	// decodes the next instruction from memory and executes it
	decodeAndExecute := func() {
		// fetch the next opcode based on the program counter
		opcode := Opcode(cpu.Memory[cpu.PC]<<8 | cpu.Memory[cpu.PC+1])

		// decode and execute the opcode
		switch opcode & 0xF000 {
		case 0x2000:
			// 2nnn - CALL addr
			// Call subroutine at nnn.
			//
			// The interpreter increments the stack pointer, then puts the current PC on the top of the stack.
			// The PC is then set to nnn.
			cpu.Stack[cpu.SP] = cpu.PC
			cpu.SP += 1

			cpu.PC = uint16(opcode & 0x0FFF)

		case 0x0004:
			// 8xy4 - ADD Vx, Vy
			// Set Vx = Vx + Vy, set VF = carry.
			//
			// The values of Vx and Vy are added together. If the result is greater than 8 bits (i.e., > 255,) VF is
			// set to 1, otherwise 0. Only the lowest 8 bits of the result are kept, and stored in Vx.
			r1 := cpu.V[(opcode&0x00F0)>>4]
			r2 := cpu.V[(opcode&0x0F00)>>8]

			if r1 > (0xFF - r2) {
				cpu.V[0xF] = 1 // carry
			} else {
				cpu.V[0xF] = 0
			}
			cpu.V[r2] += cpu.V[r1]

			cpu.PC += 2

		case 0xA000:
			// Annn - LD I, addr
			// Set I = nnn.
			//
			// The value of register I is set to nnn.
			cpu.I = uint16(opcode & 0x0FFF)
			cpu.PC += 2

		case 0x0033:
			// Fx33 - LD B, Vx
			// Store BCD representation of Vx in memory locations I, I+1, and I+2.
			//
			// The interpreter takes the decimal value of Vx, and places the hundreds digit in memory at location in I,
			// the tens digit at location I+1, and the ones digit at location I+2.
			r1 := cpu.V[(opcode&0x0F00)>>8]
			cpu.Memory[cpu.I] = cpu.V[r1] / 100
			cpu.Memory[cpu.I+1] = (cpu.V[r1] / 10) % 10
			cpu.Memory[cpu.I+2] = (cpu.V[r1] % 100) % 10

			cpu.PC += 2

		case 0xD000:
			// Dxyn - DRW Vx, Vy, nibble
			// Display n-byte sprite starting at memory location I at (Vx, Vy), set VF = collision.
			//
			// The interpreter reads n bytes from memory, starting at the address stored in I.
			// These bytes are then displayed as sprites on screen at coordinates (Vx, Vy).
			//
			// Sprites are XORed onto the existing screen. If this causes any pixels to be erased, VF is set to 1,
			// otherwise it is set to 0. If the sprite is positioned so part of it is outside the coordinates of the display,
			// it wraps around to the opposite side of the screen. See instruction 8xy3 for more information on XOR,
			// and section 2.4, Display, for more information on the Chip-8 screen and sprites.
			r1 := cpu.V[(opcode&0x00F0)>>4]
			r2 := cpu.V[(opcode&0x0F00)>>8]

			x := uint16(cpu.V[r1])
			y := uint16(cpu.V[r2])

			height := uint16(opcode & 0x000F)

			for yline := uint16(0); yline < height; yline++ {
				pixel := cpu.Memory[cpu.I+yline]

				for xline := uint16(0); xline < 8; xline++ {
					if (pixel & (0x80 >> xline)) != 0 {
						if cpu.Pixels[x+xline+((y+yline)*64)] == 1 {
							cpu.V[0xF] = 1
						}

						cpu.Pixels[x+xline+((y+yline)*64)] ^= 1
					}
				}
			}

			cpu.DrawFlag = true
			cpu.PC += 2

		default:
			println("Unknown opcode:", opcode)
		}
	}

	decodeAndExecute()
	updateTimers()
}
