package main

import (
	"os"
	"io/ioutil"
	"github.com/veandco/go-sdl2/sdl"
	"github.com/xeusalmighty/chip8emu/chip8"
	"flag"
	"log"
)

var (
	// Command line flags and options
	filenameFlag = flag.String("filename", "programs/GAMES/PONG", "The path to the program to load in the emulator")
	widthFlag    = flag.Int("width", 1024, "The width of the emulator window")
	heightFlag   = flag.Int("height", 768, "The height of the emulator window")
)

// Entry point for the emulator
func main() {
	parseCommandLine() // parse flags

	// load a test program
	cpu := chip8.NewCpu()
	cpu.LoadProgram(read(*filenameFlag))

	run(func(renderer *sdl.Renderer) {
		cpu.NextCycle() // advance the active program by 1 cycle
		if cpu.DrawFlag { // render the chip8 display, if it's been updated
			// clear the surface
			renderer.SetDrawColor(0, 0, 0, 0)
			renderer.Clear()
			// render each pixel in the pixel buffer
			// TODO: consider using a bitmap or surface here, instead?
			renderer.SetDrawColor(255, 255, 255, 255)
			for x := 0; x < 64; x++ {
				for y := 0; y < 32; y++ {
					if cpu.Pixels[x*y] > 0 {
						renderer.DrawPoint(x, y)
					}
				}
			}
			// present the surface
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
		panic("Failed to create main window")
	}
	defer window.Destroy()

	// create the main renderer
	renderer, err := sdl.CreateRenderer(window, -1, sdl.RENDERER_ACCELERATED)
	if err != nil {
		panic("Failed to create main renderer")
	}
	defer renderer.Destroy()

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
		// execute the next frame if we don't have any further events to process
		nextFrame(renderer)
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
