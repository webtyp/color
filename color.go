// Package color owns colour for the tinywasm framework: how a colour is
// written, how it is read, and how two of them relate.
//
// It carries no build tag, so it crosses into WASM. That is deliberate: a
// colour is a value the frontend genuinely handles, and the package holds
// nothing else — no CSS, no PDF, no files.
package color

import (
	. "github.com/tinywasm/fmt"
)

// Color is a CSS-style hex string: "#RRGGBB" or "#RGB".
//
// It stays a string on purpose. Colours are authored as hex everywhere in this
// ecosystem — in a design token, in a document theme, in a stylesheet — so a
// struct would force a conversion at every boundary and buy nothing.
type Color string

// RGB decomposes the colour into its three 0-255 channels.
//
// The empty colour is not an error: it resolves to black, which is what an
// unset field should paint. Anything else that is neither "#RGB" nor
// "#RRGGBB" is a defect in the caller and is reported.
func (c Color) RGB() (r, g, b int, err error) {
	s := string(c)
	if s == "" {
		return 0, 0, 0, nil
	}
	if s[0] == '#' {
		s = s[1:]
	}

	switch len(s) {
	case 3:
		// "#abc" is shorthand for "#aabbcc": each digit doubles.
		return c.channels(s[0:1]+s[0:1], s[1:2]+s[1:2], s[2:3]+s[2:3])
	case 6:
		return c.channels(s[0:2], s[2:4], s[4:6])
	}
	return 0, 0, 0, Errf("color", string(c), "invalid")
}

func (c Color) channels(rs, gs, bs string) (r, g, b int, err error) {
	if r, err = Convert(rs).Int(16); err != nil {
		return 0, 0, 0, Errf("color", string(c), "invalid")
	}
	if g, err = Convert(gs).Int(16); err != nil {
		return 0, 0, 0, Errf("color", string(c), "invalid")
	}
	if b, err = Convert(bs).Int(16); err != nil {
		return 0, 0, 0, Errf("color", string(c), "invalid")
	}
	return r, g, b, nil
}
