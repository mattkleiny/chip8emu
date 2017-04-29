package chip8

type Opcode uint16 // 35 possible opcodes, which are all two bytes long

// The central processing unit of the chip 8 system
type CPU struct {
	Memory    [4096]byte // The Chip 8 has fixed 4K memory in total
	Registers [16]byte   // 15 8-bit general purpose registers (V0, V1 through to VE). The 16th is the carry flag
	I, PC     uint16     // An index register I, and a program counter PC which can have values from 0x000 to 0xFFF
	Stack     [16]uint16 // The currently executing instruction; references the program counter
	SP        uint16     // The stack pointer

	KeyState [16]byte      // The state of each individual key (up/down)
	Pixels   [64 * 32]byte // 64 * 32 pixels (2048 total pixels)

	DelayTimer, SoundTimer byte // When above zero, they count down to zero. Counting occurs at 60hz
}

// The font-set for the chip 8 system
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

// Re-initializes the CPU
func (cpu *CPU) Initialize() {
	cpu.PC = 0x200 // program counter starts at 0x200
	cpu.I = 0      // reset index register
	cpu.SP = 0     // reset stack pointer

	// TODO: clear display
	// TODO: clear stack
	// TODO: clear registers
	// TODO: clear memory

	// load the font-set
	for i := 0; i < len(FontSet); i++ {
		cpu.Memory[i] = FontSet[i]
	}
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

	// fetch the next opcode based on the program counter
	opcode := Opcode(cpu.Memory[cpu.PC]<<8 | cpu.Memory[cpu.PC+1])

	// decode the opcode
	switch opcode & 0xF000 {
	case 0xA000: // ANNN: Sets I to the address NNN
		cpu.I = uint16(opcode & 0x0FFF)
		cpu.PC += 2

	case 0x0000:
		switch opcode & 0x000F {
		default:
			println("Unknown opcode: ", opcode)
		}

	default:
		println("Unknown opcode: ", opcode)
	}

	updateTimers()
}
