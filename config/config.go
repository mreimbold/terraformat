package config

// Config controls which formatting rules are applied.
type Config struct {
	EnforceBlockOrder      bool
	EnforceAttributeOrder  bool
	EnforceTopLevelSpacing bool
	EnsureEOFNewline       bool
}

// Default returns the default formatting configuration.
func Default() Config {
	return Config{
		EnforceBlockOrder:      true,
		EnforceAttributeOrder:  true,
		EnforceTopLevelSpacing: true,
		EnsureEOFNewline:       true,
	}
}
