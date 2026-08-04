# Specifications: Color Piece API

The package `color` defines the following public surface:

## Types

### `type Color string`
Represents a color token, usually written as a hex-encoded string (with optional hash prefix and optional alpha channel) or standard color keywords.

## Methods

### `func (c Color) RGB() (r, g, b int, err error)`
Decomposes the color into its 0-255 channels, discarding any alpha channel.

### `func (c Color) RGBA() (r, g, b, a int, err error)`
Decomposes the color into its four 0-255 channels. Correctly handles:
- `#RGB` / `#RGBA` / `#RRGGBB` / `#RRGGBBAA`
- The `transparent` keyword (resolves to `(0, 0, 0, 0)`)
- Unset/empty colors (resolves to opaque black `(0, 0, 0, 255)`)

### `func (c Color) Luminance() float64`
Returns the WCAG relative luminance of the color (from 0.0 to 1.0).

### `func (c Color) Over(backdrop Color) Color`
Blends this translucent color over an opaque background color using standard Porter-Duff alpha compositing.

## Functions

### `func Contrast(a, b Color) float64`
Calculates the WCAG contrast ratio between two colors. Output ranges from 1.0 (identical) to 21.0 (perfect contrast).

### `func Mix(a, b Color, pct float64) Color`
Interpolates two colors in the Oklab color space based on `pct` (from `0.0` to `1.0`), applying premultiplied alpha correctly during blending. Returns the result as a hex string.
