package chip8

type Opcode uint16 // 35 possible opcodes, which are all two bytes long

// The combined chip 8 system
type System struct {
	CPU      CPU
	Memory   Memory
	Keypad   Keypad
	Graphics Graphics
	Audio    Audio
}

// The central processing unit of the chip 8 system
type CPU struct {
	Registers [16]byte   // 15 8-bit general purpose registers (V0, V1 through to VE). The 16th is the carry flag
	I, PC     uint16     // An index register I, and a program counter PC which can have values from 0x000 to 0xFFF
	Stack     [16]uint16 // The currently executing instruction; references the program counter
	SP        uint16     // The stack pointer
}

// The memory bank of the chip 8 system
type Memory struct {
	Memory  [4096]byte // The Chip 8 has fixed 4K memory in total
	System  []byte     // The core system memory; a slice of the system memory over 0x000-0x1FF
	Font    []byte     // The font set; A slice of the system memory over 0x050-0x0A0
	Working []byte     // The working ROM and RAM; a slice of the system memory over 0x200-0xFFF
}

// The keypad on the chip 8 system
type Keypad struct {
	KeyState [16]byte // The state of each individual key (up/down)
}

// The graphics array on the chip 8 system
type Graphics struct {
	Pixels [64 * 32]byte // 64 * 32 pixels (2048 total pixels)
}

// The audio controller on the chip 8 system
type Audio struct {
	DelayTimer, SoundTimer byte // When above zero, they count down to zero. Counting occurs at 60hz
}
