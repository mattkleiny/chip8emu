package main

import (
	"os"
	"io/ioutil"
	"github.com/veandco/go-sdl2/sdl"
	"github.com/xeusalmighty/chip8emu/chip8"
)

// Entry point for the emulator
func main() {
	// load our test program
	cpu := chip8.NewCpu()
	cpu.LoadProgram(read("programs/GAMES/PONG"))

	nextFrame := func() {
		cpu.NextCycle()
	}

	run(100, 100, 800, 600, nextFrame)
}

// Bootstraps and executes the application via SDL
func run(x, y, w, h int, nextFrame func()) {
	sdl.Init(sdl.INIT_VIDEO | sdl.INIT_AUDIO)

	// create a window with OpenGL
	window, err := sdl.CreateWindow("chip8emu", x, y, w, h, sdl.WINDOW_OPENGL)
	if err != nil {
		panic("Failed to create main window")
	}
	defer window.Destroy()

	// runs the SDL event loop and calls the given callback function after event processing each cycle
	running := true
	for running == true {
		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch e := event.(type) {
			case *sdl.QuitEvent:
				running = false
			case *sdl.KeyDownEvent:
				// exit if 'esc' is pressed
				if e.Keysym.Sym == sdl.K_ESCAPE {
					running = false
				}
			}
		}
		nextFrame()   // execute the next frame
		sdl.Delay(16) // don't eat the cpu
	}

	sdl.Quit()
}

// Reads all of the bytes from the given file
func read(filename string) []byte {
	file, err := os.Open(filename)

	if err != nil {
		panic("Failed to open file: " + filename + " for reading")
	}

	bytes, err := ioutil.ReadAll(file)
	if err != nil {
		panic("An error occurred whilst reading bytes from file " + filename)
	}

	return bytes
}
