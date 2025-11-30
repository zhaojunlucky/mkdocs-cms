package md

var (
	DirectionRead  = "read"
	DirectionWrite = "write"
	DirectionBoth  = "both"
)

type MDConfig struct {
	CodeBlockTransforms []CodeBlockTransform `yaml:"code_block_transforms"`
}

type CodeBlockTransform struct {
	FromLang  string `yaml:"from_lang"`
	ToLang    string `yaml:"to_lang"`
	Direction string `yaml:"direction"` // "read": replace ToLang->FromLang when reading, "write": replace FromLang->ToLang when writing, "both": apply both transformations
	Enabled   *bool  `yaml:"enabled"`
}

func (c *CodeBlockTransform) CheckDirection(direction string) bool {
	if c.Direction == "" || c.Direction == DirectionBoth {
		return true
	}
	return c.Direction == direction
}

func (c *CodeBlockTransform) GetLang(direction string) (string, string) {
	if direction == DirectionRead {
		return c.ToLang, c.FromLang
	} else {
		return c.FromLang, c.ToLang
	}
}

// IsEnabled returns true if the transform is enabled, false otherwise, defaulting to true if nil
func (c *CodeBlockTransform) IsEnabled() bool {
	if c.Enabled == nil {
		return true
	} else {
		return *c.Enabled
	}
}
