// Package color owns colour for the tinywasm framework: how a colour is
// written, how it is read, and how two of them relate.
//
// It carries no build tag, so it crosses into WASM. That is deliberate: a
// colour is a value the frontend genuinely handles, and the package holds
// nothing else — no CSS, no PDF, no files.
package color

import (
	"math"

	. "github.com/tinywasm/fmt"
)

// Color is a CSS-style hex string: "#RRGGBB" or "#RGB", and supports alpha as "#RRGGBBAA" or "#RGBA", as well as "transparent".
//
// It stays a string on purpose. Colours are authored as hex everywhere in this
// ecosystem — in a design token, in a document theme, in a stylesheet — so a
// struct would force a conversion at every boundary and buy nothing.
type Color string

// RGB decomposes the colour into its three 0-255 channels, discarding any alpha channel if present.
//
// The empty colour is not an error: it resolves to black, which is what an
// unset field should paint. Anything else that is neither "#RGB", "#RGBA", "#RRGGBB", "#RRGGBBAA"
// nor "transparent" is a defect in the caller and is reported.
func (c Color) RGB() (r, g, b int, err error) {
	r, g, b, _, err = c.RGBA()
	return r, g, b, err
}

// RGBA decomposes the colour into its four 0-255 channels (red, green, blue, alpha).
//
// The empty colour is not an error: it resolves to opaque black (0, 0, 0, 255).
// The "transparent" keyword is also supported and resolves to alpha 0 (0, 0, 0, 0).
func (c Color) RGBA() (r, g, b, a int, err error) {
	s := string(c)
	if s == "" {
		return 0, 0, 0, 255, nil
	}
	if s == "transparent" {
		return 0, 0, 0, 0, nil
	}
	if s[0] == '#' {
		s = s[1:]
	}

	switch len(s) {
	case 3:
		r, g, b, err = c.channels(s[0:1]+s[0:1], s[1:2]+s[1:2], s[2:3]+s[2:3])
		return r, g, b, 255, err
	case 4:
		r, g, b, err = c.channels(s[0:1]+s[0:1], s[1:2]+s[1:2], s[2:3]+s[2:3])
		if err != nil {
			return 0, 0, 0, 0, err
		}
		var aVal int
		if aVal, err = Convert(s[3:4]+s[3:4]).Int(16); err != nil {
			return 0, 0, 0, 0, Errf("color", string(c), "invalid")
		}
		return r, g, b, aVal, nil
	case 6:
		r, g, b, err = c.channels(s[0:2], s[2:4], s[4:6])
		return r, g, b, 255, err
	case 8:
		r, g, b, err = c.channels(s[0:2], s[2:4], s[4:6])
		if err != nil {
			return 0, 0, 0, 0, err
		}
		var aVal int
		if aVal, err = Convert(s[6:8]).Int(16); err != nil {
			return 0, 0, 0, 0, Errf("color", string(c), "invalid")
		}
		return r, g, b, aVal, nil
	}
	return 0, 0, 0, 0, Errf("color", string(c), "invalid")
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

// Luminance returns the WCAG relative luminance of the color.
// If the color is invalid, it returns 0.0.
func (c Color) Luminance() float64 {
	r, g, b, err := c.RGB()
	if err != nil {
		return 0.0
	}
	return relativeLuminance(r, g, b)
}

func relativeLuminance(r, g, b int) float64 {
	rs := sRGBtoLinear(float64(r) / 255.0)
	gs := sRGBtoLinear(float64(g) / 255.0)
	bs := sRGBtoLinear(float64(b) / 255.0)
	return 0.2126*rs + 0.7152*gs + 0.0722*bs
}

func sRGBtoLinear(c float64) float64 {
	if c <= 0.04045 {
		return c / 12.92
	}
	return math.Pow((c+0.055)/1.055, 2.4)
}

// Contrast computes the WCAG contrast ratio between two colors: de 1 (identicos) a 21 (negro/blanco).
func Contrast(a, b Color) float64 {
	l1 := a.Luminance()
	l2 := b.Luminance()
	if l1 > l2 {
		return (l1 + 0.05) / (l2 + 0.05)
	}
	return (l2 + 0.05) / (l1 + 0.05)
}

// Mix blends two colors in the Oklab color space.
// pct is the interpolation weight of b from 0.0 (all a) to 1.0 (all b).
func Mix(a, b Color, pct float64) Color {
	if pct < 0.0 {
		pct = 0.0
	} else if pct > 1.0 {
		pct = 1.0
	}

	rA, gA, bA, alphaA, errA := a.RGBA()
	rB, gB, bB, alphaB, errB := b.RGBA()
	if errA != nil || errB != nil {
		return ""
	}

	// sRGB to linear sRGB
	linRA := sRGBtoLinear(float64(rA) / 255.0)
	linGA := sRGBtoLinear(float64(gA) / 255.0)
	linBA := sRGBtoLinear(float64(bA) / 255.0)

	linRB := sRGBtoLinear(float64(rB) / 255.0)
	linGB := sRGBtoLinear(float64(gB) / 255.0)
	linBB := sRGBtoLinear(float64(bB) / 255.0)

	// linear sRGB to Oklab
	LA, aA, bValA := linearSRGBToOklab(linRA, linGA, linBA)
	LB, aB, bValB := linearSRGBToOklab(linRB, linGB, linBB)

	aF := float64(alphaA) / 255.0
	bF := float64(alphaB) / 255.0

	// Premultiply
	LAPre, aAPre, bValAPre := LA*aF, aA*aF, bValA*aF
	LBPre, aBPre, bValBPre := LB*bF, aB*bF, bValB*bF

	// Interpolate
	alphaMix := aF*(1.0-pct) + bF*pct
	LMixPre := LAPre*(1.0-pct) + LBPre*pct
	aMixPre := aAPre*(1.0-pct) + aBPre*pct
	bValMixPre := bValAPre*(1.0-pct) + bValBPre*pct

	var LMix, aMix, bValMix float64
	if alphaMix > 0.0 {
		LMix = LMixPre / alphaMix
		aMix = aMixPre / alphaMix
		bValMix = bValMixPre / alphaMix
	}

	// Oklab to linear sRGB
	linRMix, linGMix, linBMix := oklabToLinearSRGB(LMix, aMix, bValMix)

	// linear sRGB to sRGB
	rMix := linearToSRGB(linRMix)
	gMix := linearToSRGB(linGMix)
	bMix := linearToSRGB(linBMix)

	rInt := int(math.Round(clamp(rMix) * 255.0))
	gInt := int(math.Round(clamp(gMix) * 255.0))
	bInt := int(math.Round(clamp(bMix) * 255.0))
	aInt := int(math.Round(clamp(alphaMix) * 255.0))

	return toHex(rInt, gInt, bInt, aInt, aInt < 255)
}

// Over composes a translucent color (this color `c`) over an opaque background color `backdrop`.
// If the backdrop has alpha, it still blends them using the standard Porter-Duff Over operator.
func (c Color) Over(backdrop Color) Color {
	rFG, gFG, bFG, aFG, errFG := c.RGBA()
	rBG, gBG, bBG, aBG, errBG := backdrop.RGBA()
	if errFG != nil || errBG != nil {
		return ""
	}

	fgA := float64(aFG) / 255.0
	bgA := float64(aBG) / 255.0

	outA := fgA + bgA*(1.0-fgA)
	if outA <= 0.0 {
		return toHex(0, 0, 0, 0, true)
	}

	// Blend in standard non-linear sRGB (matching standard canvas/CSS alpha compositing)
	outR := (float64(rFG)*fgA + float64(rBG)*bgA*(1.0-fgA)) / outA
	outG := (float64(gFG)*fgA + float64(gBG)*bgA*(1.0-fgA)) / outA
	outB := (float64(bFG)*fgA + float64(bBG)*bgA*(1.0-fgA)) / outA

	rInt := int(math.Round(clamp(outR/255.0) * 255.0))
	gInt := int(math.Round(clamp(outG/255.0) * 255.0))
	bInt := int(math.Round(clamp(outB/255.0) * 255.0))
	aInt := int(math.Round(clamp(outA) * 255.0))

	return toHex(rInt, gInt, bInt, aInt, aInt < 255)
}

func linearSRGBToOklab(r, g, b float64) (L, a, bVal float64) {
	l := 0.4122214708*r + 0.5363325363*g + 0.0514459929*b
	m := 0.2119034982*r + 0.6806995451*g + 0.1073969566*b
	s := 0.0883024619*r + 0.2817188376*g + 0.6299787005*b

	l_ := math.Cbrt(l)
	m_ := math.Cbrt(m)
	s_ := math.Cbrt(s)

	L = 0.2104542553*l_ + 0.7936177850*m_ - 0.0040720468*s_
	a = 1.9779984951*l_ - 2.4285922050*m_ + 0.4505937099*s_
	bVal = 0.0259040371*l_ + 0.7827717662*m_ - 0.8086757660*s_
	return L, a, bVal
}

func oklabToLinearSRGB(L, a, bVal float64) (r, g, b float64) {
	l_ := L + 0.3963377774*a + 0.2158037573*bVal
	m_ := L - 0.1055613458*a - 0.0638541728*bVal
	s_ := L - 0.0894841775*a - 1.2914855480*bVal

	l := l_ * l_ * l_
	m := m_ * m_ * m_
	s := s_ * s_ * s_

	r = 4.0767416621*l - 3.3077115913*m + 0.2309699292*s
	g = -1.2684380046*l + 2.6097574011*m - 0.3413193965*s
	b = -0.0041960863*l - 0.7034186147*m + 1.7076147010*s
	return r, g, b
}

func linearToSRGB(c float64) float64 {
	if c <= 0.0031308 {
		return c * 12.92
	}
	return 1.055*math.Pow(c, 1.0/2.4) - 0.055
}

func clamp(v float64) float64 {
	if v < 0.0 {
		return 0.0
	}
	if v > 1.0 {
		return 1.0
	}
	return v
}

func toHex(r, g, b, a int, includeAlpha bool) Color {
	const hexChars = "0123456789abcdef"
	if includeAlpha {
		buf := make([]byte, 9)
		buf[0] = '#'
		buf[1] = hexChars[r>>4]
		buf[2] = hexChars[r&15]
		buf[3] = hexChars[g>>4]
		buf[4] = hexChars[g&15]
		buf[5] = hexChars[b>>4]
		buf[6] = hexChars[b&15]
		buf[7] = hexChars[a>>4]
		buf[8] = hexChars[a&15]
		return Color(buf)
	} else {
		buf := make([]byte, 7)
		buf[0] = '#'
		buf[1] = hexChars[r>>4]
		buf[2] = hexChars[r&15]
		buf[3] = hexChars[g>>4]
		buf[4] = hexChars[g&15]
		buf[5] = hexChars[b>>4]
		buf[6] = hexChars[b&15]
		return Color(buf)
	}
}
