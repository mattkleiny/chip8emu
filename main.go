package main

import (
	"github.com/veandco/go-sdl2/sdl"
	"github.com/xeusalmighty/chip8emu/chip8"
)

// Entry point for the emulator
func main() {
	system := new(chip8.CPU)
	system.Initialize()

	processor := func() {
		system.NextCycle()
	}

	run(100, 100, 800, 600, processor)
}

// Bootstraps and executes the application via SDL
func run(x, y, w, h int, frameHandler func()) {
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
			switch event.(type) {
			case *sdl.QuitEvent:
				running = false
			}
		}
		frameHandler() // execute the next frame
		sdl.Delay(16)  // don't eat the cpu
	}

	sdl.Quit()
}
