// Copyright 2017, the project authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE.md file.

package chip8

type Keycode byte // Our keycode representation.

// A keypad implementation for the interpreter.
type Keypad struct {
	pressed  chan Keycode     // A channel of key down events.
	released chan Keycode     // A channel of key up events.
	states   map[Keycode]bool // A map of keys, and their state
}

// Builds a new default keypad.
func NewKeypad() *Keypad {
	const bufferSize = 1000
	keypad := &Keypad{
		pressed:  make(chan Keycode, bufferSize),
		released: make(chan Keycode, bufferSize),
		states:   make(map[Keycode]bool),
	}
	return keypad
}

// Notifies the given key was pressed.
func (keypad *Keypad) Press(key Keycode) {
	keypad.pressed <- key
	keypad.states[key] = true
}

// Notifies the given key was released.
func (keypad *Keypad) Release(key Keycode) {
	keypad.released <- key
	keypad.states[key] = false
}

// Determines if the given key is currently pressed.
func (keypad *Keypad) IsPressed(key Keycode) bool {
	return keypad.states[key]
}

// Blocks and reads a key from the keypad.
func (keypad *Keypad) Read() (Keycode, error) {
	select {
	case key := <-keypad.pressed:
		return key, nil

	case key := <-keypad.pressed:
		return key, nil
	}
}

// Closes the keypad by releasing it's channels.
func (keypad *Keypad) Close() {
	close(keypad.pressed)
	close(keypad.released)
}
