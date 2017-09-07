// Copyright 2017, the project authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE.md file.

package main

import (
	"flag"
	"github.com/veandco/go-sdl2/sdl"
	"github.com/xeusalmighty/chip8emu/chip8"
	"io/ioutil"
	"log"
)

var ( // Command line flags and arguments
	filenameFlag  = flag.String("filename", "programs/GAMES/MERLIN", "The path to the program to load into the interpreter")
	widthFlag     = flag.Int("width", 1024, "The width of the window")
	heightFlag    = flag.Int("height", 768, "The height of the window")
	frequencyFlag = flag.Uint("frequency", 60, "The frequency, in hertz, to run the processor at")
)

// the singleton chip 8 cpu
var cpu = chip8.NewCPU()

// A mapping of SDL to Chip-8 key codes.
var keycodes = map[sdl.Keycode]chip8.Keycode{
	sdl.K_1: 0x01, sdl.K_2: 0x02, sdl.K_3: 0x03, sdl.K_4: 0x0C,
	sdl.K_q: 0x04, sdl.K_w: 0x05, sdl.K_e: 0x06, sdl.K_r: 0x0D,
	sdl.K_a: 0x07, sdl.K_s: 0x08, sdl.K_d: 0x09, sdl.K_f: 0x0E,
	sdl.K_z: 0x0A, sdl.K_x: 0x00, sdl.K_c: 0x0B, sdl.K_v: 0x0F,
}

// Entry point for the interpreter
func main() {
	parseCommandLine()

	// load a test program and start it executing in the background
	cpu.LoadProgram(readFile(*filenameFlag))
	go cpu.RunAtFrequency(*frequencyFlag)

	// start winding up SDL
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

	// create a texture mimicking the default dimensions of the chip8 display
	texture, err := renderer.CreateTexture(sdl.PIXELFORMAT_RGBA8888, sdl.TEXTUREACCESS_TARGET, chip8.Width, chip8.Height)
	if err != nil {
		log.Fatal("Failed to create main texture. ", err)
	}
	defer texture.Destroy()

	// run the main event loop
	running := true
	for running {
		// process incoming events
		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch e := event.(type) {
			case *sdl.QuitEvent:
				running = false

			case *sdl.KeyDownEvent:
				// exit if escape is pressed
				if e.Keysym.Sym == sdl.K_ESCAPE {
					running = false
				}
				cpu.Keypad.Press(keycodes[e.Keysym.Sym])

			case *sdl.KeyUpEvent:
				cpu.Keypad.Release(keycodes[e.Keysym.Sym])
			}
		}

		renderer.SetRenderTarget(texture)
		renderer.SetDrawColor(0, 0, 0, 0)
		renderer.Clear()

		renderer.SetDrawColor(255, 255, 255, 255)
		for x := 0; x < chip8.Width; x++ {
			for y := 0; y < chip8.Height; y++ {
				// draw active pixels
				if cpu.Pixels.GetPixel(x, y) > 0 {
					renderer.DrawPoint(x, y)
				}
			}
		}

		renderer.Present()
		renderer.SetRenderTarget(nil)
		renderer.Copy(texture, nil, nil)

		// don't eat the cpu
		sdl.Delay(1000 / 60)
	}

	sdl.Quit()
}

// Reads all of the bytes from the given file.
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

	if *frequencyFlag == 0 {
		flag.Usage()
		log.Fatal("A valid frequency was expected")
	}
}
