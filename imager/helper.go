package main

import (
	"fmt"
	"image/color"
	"log/slog"
	"os"
	"slices"
	"strings"

	"golang.org/x/image/math/fixed"
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

func getTextPoint(position string, text string, size int) fixed.Point26_6 {
	var ret fixed.Point26_6

	poscomp := strings.Split(position, "-")
	if len(poscomp) != 2 {
		slog.Error("failed to parse label position", "error", "incorrect format")
		os.Exit(1)
	}

	if !slices.Contains([]string{"top", "bottom"}, poscomp[0]) {
		slog.Error("failed to parse label position", "error", "pick either top or bottom")
		os.Exit(1)
	}

	if !slices.Contains([]string{"left", "right"}, poscomp[1]) {
		slog.Error("failed to parse label position", "error", "pick either left or right")
		os.Exit(1)
	}

	if poscomp[0] == "top" {
		ret.Y = fixed.I(13) /* 13 is font height */
	} else { /* Bottom */
		ret.Y = fixed.I(size - 13 + 5) /* 13 is the font height, 5 is the padding */
	}

	if poscomp[1] == "left" {
		ret.X = fixed.I(5)
	} else { /* Right */
		ret.X = fixed.I(size - 5 - 7*len(text)) /* 5 is padding, 7 is the font width */
	}

	return ret
}
