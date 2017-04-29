package main

import "github.com/veandco/go-sdl2/sdl"

func main() {
	sdl.Init(sdl.INIT_VIDEO | sdl.INIT_AUDIO)

	window, err := sdl.CreateWindow("chip8emu", 100, 100, 800, 600, sdl.WINDOW_OPENGL)
	if err != nil {
		panic("Failed to create main window")
	}
	defer window.Destroy()

	// TODO: run the main loop

	sdl.Quit()
}
