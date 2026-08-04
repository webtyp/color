package color

import (
	"math"
	"testing"
)

// The three shapes a caller can write, and the one that is not an error.
func TestRGB(t *testing.T) {
	for _, c := range []struct {
		in      Color
		r, g, b int
		wantErr bool
		why     string
	}{
		{"#1E3C78", 30, 60, 120, false, "the six-digit form"},
		{"1E3C78", 30, 60, 120, false, "the hash is optional"},
		{"#abc", 170, 187, 204, false, "shorthand doubles each digit"},
		{"", 0, 0, 0, false, "an unset colour paints black, it is not a defect"},
		{"#12345", 0, 0, 0, true, "five digits is neither form"},
		{"#GGGGGG", 0, 0, 0, true, "not hexadecimal"},
	} {
		r, g, b, err := c.in.RGB()
		if (err != nil) != c.wantErr {
			t.Errorf("%q: err = %v, wantErr = %v — %s", c.in, err, c.wantErr, c.why)
			continue
		}
		if !c.wantErr && (r != c.r || g != c.g || b != c.b) {
			t.Errorf("%q: got (%d,%d,%d), want (%d,%d,%d) — %s",
				c.in, r, g, b, c.r, c.g, c.b, c.why)
		}
	}
}

// Shorthand and its long form must decompose identically, or the two spellings
// of the same colour would paint differently.
func TestShorthandMatchesLongForm(t *testing.T) {
	sr, sg, sb, err := Color("#abc").RGB()
	if err != nil {
		t.Fatalf("shorthand: %v", err)
	}
	lr, lg, lb, err := Color("#aabbcc").RGB()
	if err != nil {
		t.Fatalf("long form: %v", err)
	}
	if sr != lr || sg != lg || sb != lb {
		t.Errorf("#abc gave (%d,%d,%d) but #aabbcc gave (%d,%d,%d)", sr, sg, sb, lr, lg, lb)
	}
}

// TestRGBA covers alpha parsing, #RGBA and #RRGGBBAA formats, and the transparent keyword.
func TestRGBA(t *testing.T) {
	for _, c := range []struct {
		in         Color
		r, g, b, a int
		wantErr    bool
		why        string
	}{
		{"#1E3C78FF", 30, 60, 120, 255, false, "eight-digit form fully opaque"},
		{"#1E3C7880", 30, 60, 120, 128, false, "eight-digit form semi-transparent"},
		{"#1234", 17, 34, 51, 68, false, "four-digit shorthand"},
		{"transparent", 0, 0, 0, 0, false, "transparent keyword"},
		{"#1234567", 0, 0, 0, 0, true, "seven digits is invalid"},
	} {
		r, g, b, a, err := c.in.RGBA()
		if (err != nil) != c.wantErr {
			t.Errorf("%q: err = %v, wantErr = %v — %s", c.in, err, c.wantErr, c.why)
			continue
		}
		if !c.wantErr && (r != c.r || g != c.g || b != c.b || a != c.a) {
			t.Errorf("%q: got (%d,%d,%d,%d), want (%d,%d,%d,%d) — %s",
				c.in, r, g, b, a, c.r, c.g, c.b, c.a, c.why)
		}
	}
}

// TestLuminance checks sRGB relative luminance calculation.
func TestLuminance(t *testing.T) {
	// White relative luminance is 1.0, black is 0.0
	whiteL := Color("#ffffff").Luminance()
	if math.Abs(whiteL-1.0) > 1e-5 {
		t.Errorf("Luminance of white: got %f, want 1.0", whiteL)
	}

	blackL := Color("#000000").Luminance()
	if math.Abs(blackL-0.0) > 1e-5 {
		t.Errorf("Luminance of black: got %f, want 0.0", blackL)
	}

	// Mid gray #808080 luminance
	grayL := Color("#808080").Luminance()
	// c = 128.0/255.0 = 0.50196078
	// lin = math.Pow((c+0.055)/1.055, 2.4) = 0.2158605
	if math.Abs(grayL-0.2158605) > 1e-5 {
		t.Errorf("Luminance of #808080: got %f, want ~0.2158605", grayL)
	}
}

// TestContrast checks WCAG contrast ratio calculation.
func TestContrast(t *testing.T) {
	// Identical colors contrast must be exactly 1.0
	c1 := Contrast(Color("#ffffff"), Color("#ffffff"))
	if math.Abs(c1-1.0) > 1e-5 {
		t.Errorf("Contrast of white with white: got %f, want 1.0", c1)
	}

	// Black and white contrast must be exactly 21.0
	c21 := Contrast(Color("#ffffff"), Color("#000000"))
	if math.Abs(c21-21.0) > 1e-5 {
		t.Errorf("Contrast of white and black: got %f, want 21.0", c21)
	}

	// Verify custom contrast pair from plan: #e8a33d on #1C1C1E ≈ 7.89
	cReal := Contrast(Color("#e8a33d"), Color("#1C1C1E"))
	if math.Abs(cReal-7.89) > 0.05 {
		t.Errorf("Contrast of #e8a33d on #1C1C1E: got %f, want ~7.89", cReal)
	}
}

// TestMix checks blending of colors in Oklab color space.
func TestMix(t *testing.T) {
	// 50% blend of #ff0000 and #0000ff should give oklab(0.539974 0.0962086 -0.0928316) -> sRGB ~140 83 162 => #8c53a2
	mix := Mix(Color("#ff0000"), Color("#0000ff"), 0.5)
	if mix != "#8c53a2" {
		t.Errorf("Oklab 50%% mix of red and blue: got %s, want #8c53a2", mix)
	}

	// 30% mix of #e8a33d and #1C1C1E (70% #e8a33d, 30% #1C1C1E) -> sRGB ~165 119 58 => #a5773a
	mix30 := Mix(Color("#e8a33d"), Color("#1C1C1E"), 0.3)
	if mix30 != "#a5773a" {
		t.Errorf("Oklab 30%% mix: got %s, want #a5773a", mix30)
	}

	// Mix with transparent (alpha 85% transparent, 15% #e8a33d) => oklab(0.764905 0.0410493 0.133744 / 0.15)
	// => sRGB ~ 232 163 61, alpha 38 => #e8a33d26
	mixTrans := Mix(Color("#e8a33d"), Color("transparent"), 0.85)
	if mixTrans != "#e8a33d26" {
		t.Errorf("Oklab mix with transparent: got %s, want #e8a33d26", mixTrans)
	}
}

// TestOver checks alpha compositing / overlaying.
func TestOver(t *testing.T) {
	// Full translucent gold #e8a33d over background #1C1C1E at 15% alpha
	// fg = #e8a33d26 (aFG = 38)
	// bg = #1C1C1E (rBG=28, gBG=28, bBG=30)
	// outR = 58, outG = 48, outB = 35 => #3a3023
	over := Color("#e8a33d26").Over(Color("#1C1C1E"))
	if over != "#3a3023" {
		t.Errorf("Over blend: got %s, want #3a3023", over)
	}

	// Fully transparent fg leaves bg intact
	overTrans := Color("transparent").Over(Color("#1C1C1E"))
	if overTrans != "#1c1c1e" {
		t.Errorf("Over with transparent fg: got %s, want #1c1c1e", overTrans)
	}

	// Fully opaque fg covers bg completely
	overOpaque := Color("#ffffff").Over(Color("#000000"))
	if overOpaque != "#ffffff" {
		t.Errorf("Over with opaque fg: got %s, want #ffffff", overOpaque)
	}
}
