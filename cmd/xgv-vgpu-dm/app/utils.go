package app

import (
	"fmt"
	"strings"
)

func CheckFlags(f *Flags) error {
	var missing []string
	if f.ConfigFile == "" {
		missing = append(missing, "config-file")
	}
	if len(missing) > 0 {
		return fmt.Errorf("missing required flags '%v'", strings.Join(missing, ", "))
	}
	return nil
}
