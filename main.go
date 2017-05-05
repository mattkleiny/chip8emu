// Copyright © 2017 Matthew Kleinschafer
//
// Permission is hereby granted, free of charge, to any person
// obtaining a copy of this software and associated documentation
// files (the “Software”), to deal in the Software without
// restriction, including without limitation the rights to use,
// copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the
// Software is furnished to do so, subject to the following
// conditions:
//
// The above copyright notice and this permission notice shall be
// included in all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED “AS IS”, WITHOUT WARRANTY OF ANY KIND,
// EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES
// OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND
// NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT
// HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY,
// WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING
// FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
// OTHER DEALINGS IN THE SOFTWARE.

package main

import (
	"flag"
	"io/ioutil"
	"log"
	"github.com/veandco/go-sdl2/sdl"
	"github.com/xeusalmighty/chip8emu/chip8"
)

var ( // Command line flags and arguments
	filenameFlag  = flag.String("filename", "programs/GAMES/PONG", "The path to the program to load into the interpreter")
	widthFlag     = flag.Int("width", 1024, "The width of the window")
	heightFlag    = flag.Int("height", 768, "The height of the window")
	frequencyFlag = flag.Uint("frequency", 60, "The frequency, in hertz, to run the processor at")
)

// the singleton chip 8 cpu
var cpu = chip8.NewCPU()

// map of runes to the associated flag in the cpu
var keypadLookup = map[sdl.Scancode]*bool{
	scancode('Q'): &cpu.Keypad[0],
	scancode('W'): &cpu.Keypad[1],
	scancode('E'): &cpu.Keypad[2],
	scancode('R'): &cpu.Keypad[3],
	scancode('A'): &cpu.Keypad[4],
	scancode('S'): &cpu.Keypad[5],
	scancode('D'): &cpu.Keypad[6],
	scancode('F'): &cpu.Keypad[7],
	scancode('Z'): &cpu.Keypad[8],
	scancode('X'): &cpu.Keypad[9],
	scancode('C'): &cpu.Keypad[10],
	scancode('V'): &cpu.Keypad[11],
	scancode('T'): &cpu.Keypad[12],
	scancode('G'): &cpu.Keypad[13],
	scancode('B'): &cpu.Keypad[14],
	scancode('N'): &cpu.Keypad[15],
}

// Retrieves the scan code for the given rune.
func scancode(rune rune) sdl.Scancode {
	return sdl.GetScancodeFromName(string(rune))
}

// Entry point for the interpreter
func main() {
	parseCommandLine()

	// load a test program
	cpu.LoadProgram(readFile(*filenameFlag))

	// run the cpu at a fixed frequency
	go cpu.RunAtFrequency(*frequencyFlag)

	// run the main event loop
	run(func(renderer *sdl.Renderer) {
		updateDisplay(renderer)
		updateKeypad()
	})
}

// Handles updating the window display via SDL.
func updateDisplay(renderer *sdl.Renderer) {
	// clear the display
	renderer.SetDrawColor(0, 0, 0, 0)
	renderer.Clear()

	// render each pixel in the bitmap
	renderer.SetDrawColor(255, 255, 255, 255)
	for x := 0; x < chip8.Width-1; x++ {
		for y := 0; y < chip8.Height-1; y++ {
			// draw active pixels
			if cpu.Pixels.GetPixel(x, y) > 0 {
				renderer.DrawPoint(x, y)
			}
		}
	}

	// present the display
	renderer.Present()
}

// Handles input translation to the chip 8 keypad.
func updateKeypad() {
	state := sdl.GetKeyboardState()
	// check each key in our keyboard map and see if it's pressed
	// if it is, update the associated flag in the cpu
	for scancode, flag := range keypadLookup {
		if state[scancode] == 1 {
			*flag = true
		} else {
			*flag = false
		}
	}
}

// Bootstraps and executes the application via SDL
func run(update func(renderer *sdl.Renderer)) {
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
				if e.Keysym.Sym == sdl.K_ESCAPE {
					running = false
				}
			}
		}

		// render to our display texture
		renderer.SetRenderTarget(texture)

		// execute the next frame if we don't have any further events to process
		update(renderer)

		// upscale and copy the texture back to the window
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
