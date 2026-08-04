package color

import "testing"

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
