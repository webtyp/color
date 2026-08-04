# Architecture: Color Piece

This module provides WCAG-compliant color utilities for the tinywasm ecosystem. It is designed to compile cleanly to WebAssembly (`GOOS=js GOARCH=wasm`) without any build tags and operates purely with string-based design tokens (`Color`).

## Core Responsibilities

- **Parsing & Decomposing**: Handles conversion from hex-encoded formats (`#RGB`, `#RGBA`, `#RRGGBB`, `#RRGGBBAA`) and keywords (`transparent`) into raw RGB/RGBA channels.
- **Relative Luminance**: Implements standard sRGB-to-linear conversion curves and coefficients defined by the WCAG spec to measure relative luminance.
- **Contrast Ratios**: Computes exact contrast ratios between design tokens for accessibility audits.
- **Oklab Color Mixing**: Interpolates colors correctly in the perceptual Oklab color space (applying premultiplied alpha where applicable), preventing muddy midtones or incorrect hues during interpolation.
- **Compositing**: Blends translucent foregrounds over opaque backdrops using the standard Porter-Duff `Over` operator.

## Architecture Guidelines

This piece is strictly focused on color mathematics and representations. To maintain a clean and minimal surface:
- No layout, page sizes, or CSS stylesheet emitter code can exist here.
- No third-party map data structures are used, preventing excessive binary bloat in WebAssembly outputs.
