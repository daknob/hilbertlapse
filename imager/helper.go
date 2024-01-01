package main

import (
	"fmt"
	"image/color"
	"log/slog"
	"os"
)

func parseColor(in string) color.RGBA {
	var ret color.RGBA
	ret.A = 255

	n, err := fmt.Sscanf(in, "#%02x%02x%02x", &ret.R, &ret.G, &ret.B)
	if n != 3 || err != nil {
		slog.Error("failed to parse color", "parsed", n, "error", err)
		os.Exit(1)
	}

	return ret
}
