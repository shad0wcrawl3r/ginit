package config

import (
	"fmt"
	"unicode"
)

type Config struct {
	PackageName string
	Template    string
}

func (c Config) Validate() error {
	s := c.PackageName
	if s == "" {
		return fmt.Errorf("invalid package name: empty")
	}
	if !unicode.IsLetter(rune(s[0])) {
		return fmt.Errorf("invalid package name: must start with a letter")
	}
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' {
			continue
		}
		return fmt.Errorf("invalid package name: invalid character %q", r)
	}
	return nil
}
