// logging.go

package main

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/jwalton/go-supportscolor"
)

// print to stdout with color
func printColored(
	c color.Attribute,
	format string,
	a ...any,
) {
	formatted := fmt.Sprintf(format, a...)

	if supportscolor.Stdout().SupportsColor { // if color is supported,
		c := color.New(c)
		_, _ = c.Printf("%s", formatted)
	} else {
		fmt.Print(formatted)
	}
}
