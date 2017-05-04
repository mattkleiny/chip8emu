package main

import (
	"flag"
	"io/ioutil"
	"log"

	"github.com/veandco/go-sdl2/sdl"
	"github.com/xeusalmighty/chip8emu/chip8"
)

var (
	filenameFlag = flag.String("filename", "programs/GAMES/PONG", "The path to the program to load in the emulator")
	widthFlag    = flag.Int("width", 1024, "The width of the emulator window")
	heightFlag   = flag.Int("height", 768, "The height of the emulator window")
)

// Entry point for the emulator
func main() {
	parseCommandLine()

	// load a test program
	cpu := chip8.NewCPU()
	cpu.LoadProgram(readFile(*filenameFlag))

	run(func(renderer *sdl.Renderer) {
		cpu.NextCycle() // advance the active program by 1 cycle

		if cpu.DrawFlag { // render the chip8 display, if it's been updated
			// clear the surface
			renderer.SetDrawColor(0, 0, 0, 0)
			renderer.Clear()

			// render each pixel in the pixel buffer
			// TODO: consider using a bitmap or surface here, instead?
			renderer.SetDrawColor(255, 255, 255, 255)
			for x := 0; x < chip8.Width; x++ {
				for y := 0; y < chip8.Height; y++ {
					// TODO: draw a pixel
				}
			}

			// present the display
			renderer.Present()
		}
	})
}

// Bootstraps and executes the application via SDL
func run(nextFrame func(renderer *sdl.Renderer)) {
	sdl.Init(sdl.INIT_VIDEO)

	// create the main window
	window, err := sdl.CreateWindow("chip8emu", 100, 100, *widthFlag, *heightFlag, sdl.WINDOW_SHOWN)
	if err != nil {
		log.Fatal("Failed to create main window. ", err)
	}
	defer window.Destroy()

	// create the main renderer
	renderer, err := sdl.CreateRenderer(window, -1, sdl.RENDERER_ACCELERATED)
	if err != nil {
		log.Fatal("Failed to create main renderer. ", err)
	}
	defer renderer.Destroy()

	// run the main event loop
	running := true
	for running {
		// process incoming events
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
		// execute the next frame if we don't have any further events to process
		nextFrame(renderer)
	}

	sdl.Quit()
}

// Reads all of the bytes from the given file
func readFile(filename string) []byte {
	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatal("An error occurred whilst reading file. ", err)
	}
	return bytes
}

// Parse and validate command line arguments.
func parseCommandLine() {
	flag.Parse()

	if *filenameFlag == "" {
		flag.Usage()
		log.Fatal("A valid filename was expected")
	}
	if *widthFlag == 0 {
		flag.Usage()
		log.Fatal("A valid width was expected")
	}
	if *heightFlag == 0 {
		flag.Usage()
		log.Fatal("A valid height was expected")
	}
}
