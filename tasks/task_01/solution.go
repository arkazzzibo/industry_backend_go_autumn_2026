package main

import (
	"fmt"
	"strings"
)

func greet(name string) string {
	name = strings.TrimSpace(name)

	if name == "" {
		return "Hello, World!"
	}
	return fmt.Sprintf("Hello, %s!", name)
}
