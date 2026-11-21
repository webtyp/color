package color

import "webtyp.com/input"

type colorInput struct{ input.Base }

// Input returns a native HTML color picker (<input type="color">) validated by Color.RGBA.
func Input() input.Input {
	c := &colorInput{}
	c.Letters = true
	c.Numbers = true
	c.Extra = []rune{'#'}
	c.Minimum = 7
	c.Maximum = 7
	c.InitBase("", "", "color")
	return c
}

// Validate delegates directly to Color.RGBA so colour format rules live in one place.
func (c *colorInput) Validate(value string) error {
	if value == "" {
		return nil
	}
	_, _, _, _, err := Color(value).RGBA()
	return err
}

// Clone satisfies input.Input.
func (c *colorInput) Clone(parentID, name string) input.Input {
	cp := *c
	cp.InitBase(parentID, name, "color")
	return &cp
}
