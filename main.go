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

// map of scan-codes to the associated flag in the cpu
var keypadLookup = map[sdl.Scancode]*bool{
	sdl.K_1: &cpu.Keypad[0x1],
	sdl.K_2: &cpu.Keypad[0x2],
	sdl.K_3: &cpu.Keypad[0x3],
	sdl.K_4: &cpu.Keypad[0xC],
	sdl.K_5: &cpu.Keypad[0x5],
	sdl.K_6: &cpu.Keypad[0x6],
	sdl.K_7: &cpu.Keypad[0x7],
	sdl.K_8: &cpu.Keypad[0x8],
	sdl.K_9: &cpu.Keypad[0x9],
	sdl.K_0: &cpu.Keypad[0x0],
	sdl.K_q: &cpu.Keypad[0x4],
	sdl.K_w: &cpu.Keypad[0x5],
	sdl.K_e: &cpu.Keypad[0x6],
	sdl.K_r: &cpu.Keypad[0xD],
	sdl.K_a: &cpu.Keypad[0x7],
	sdl.K_s: &cpu.Keypad[0x8],
	sdl.K_d: &cpu.Keypad[0x9],
	sdl.K_f: &cpu.Keypad[0xE],
	sdl.K_z: &cpu.Keypad[0xA],
	sdl.K_x: &cpu.Keypad[0x0],
	sdl.K_c: &cpu.Keypad[0x7],
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
				flag, ok := keypadLookup[e.Keysym.Scancode]
				if ok {
					*flag = true
				}

				// exit if escape is pressed
				if e.Keysym.Sym == sdl.K_ESCAPE {
					running = false
				}

			case *sdl.KeyUpEvent:
				flag, ok := keypadLookup[e.Keysym.Scancode]
				if ok {
					*flag = false
				}
			}
		}

		// render to our display texture
		renderer.SetRenderTarget(texture)
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
