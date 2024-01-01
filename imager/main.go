/*
 * Copyright (c) 2020 Antonios A. Chariton <daknob@daknob.net>
 *
 * Permission to use, copy, modify, and distribute this software for any
 * purpose with or without fee is hereby granted, provided that the above
 * copyright notice and this permission notice appear in all copies.
 *
 * THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES
 * WITH REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF
 * MERCHANTABILITY AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR
 * ANY SPECIAL, DIRECT, INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES
 * WHATSOEVER RESULTING FROM LOSS OF USE, DATA OR PROFITS, WHETHER IN AN
 * ACTION OF CONTRACT, NEGLIGENCE OR OTHER TORTIOUS ACTION, ARISING OUT OF
 * OR IN CONNECTION WITH THE USE OR PERFORMANCE OF THIS SOFTWARE.
 *
 */

package main

import (
	"bufio"
	"flag"
	"image"
	"image/png"
	"log/slog"
	"math"
	"math/big"
	"net/netip"
	"os"
	"strings"

	"github.com/google/hilbert"
)

const (
	IPv4Bits    = 32
	CommentChar = "#"
)

func main() {
	slog.Info("starting imager...")

	/* Accept the file to image, default from STDIN */
	inFile := flag.String("i", "/dev/stdin", "Input CSV File")

	/* Accept the PNG file name to output as, default at STDOUT */
	outFile := flag.String("o", "/dev/stdout", "Output PNG File")

	/* Accept range to image, default to 193.5.16.0/22 */
	imageRange := flag.String("r", "193.5.16.0/22", "Output range (CIDR must be even)")

	/* Colors used for up and down hosts */
	upColor := flag.String("u", "#32c832", "Color used for hosts that are up")
	downColor := flag.String("d", "#323232", "Color used for hosts that are down")
	flag.Parse()

	/* Open the file */
	slog.Info("loading file", "name", *inFile)
	cont, err := os.Open(*inFile)
	if err != nil {
		slog.Error("could not open file", "name", *inFile, "error", err)
		os.Exit(1)
	}
	slog.Info("file loaded", "name", *inFile)

	/* Parse the image range */
	prefix, err := netip.ParsePrefix(*imageRange)
	if err != nil {
		slog.Error("failed to parse image range prefix", "error", err)
		os.Exit(1)
	}
	if !prefix.Addr().Is4() {
		slog.Error("imager does not support IPv6 at the moment :(")
		os.Exit(2)
	}
	if prefix.Bits()%2 != 0 {
		slog.Error("the image range CIDR must be even (e.g. /8, /10, /24, /30, ...)")
		os.Exit(1)
	}

	/* Canonicalize prefix (e.g. 193.5.16.1/22 -> 193.5.16.0/22) */
	prefix = prefix.Masked()

	/* Calculate the range size */
	rangeSize := int(math.Pow(2, float64(IPv4Bits-prefix.Bits())))

	/* Calculate the range dimension */
	rangeSqrt := int(math.Sqrt(float64(rangeSize)))

	/* Calculate the first IP Address as a big.Int */
	baseAs4 := prefix.Addr().As4()
	rangeStart := big.NewInt(0).SetBytes(baseAs4[:])

	slog.Info("prefix parsed", "base", prefix.Addr().String(), "length", prefix.Bits(), "size", rangeSize, "grid", rangeSqrt)

	/* Create a new PNG image in memory */
	slog.Info("creating image file...")
	resultImage := image.NewRGBA(
		image.Rectangle{
			image.Point{0, 0},
			image.Point{rangeSqrt, rangeSqrt},
		},
	)

	/* Fill image with offline color */
	for i := 0; i < rangeSqrt; i++ {
		for j := 0; j < rangeSqrt; j++ {
			resultImage.Set(i, j, parseColor(*downColor))
		}
	}

	/* Create the output file, and open it for writing */
	resultFile, err := os.Create(*outFile)
	if err != nil {
		slog.Error("failed to create image file for writing", "name", *outFile, "error", err)
		os.Exit(1)
	}

	/* Create a new Hilbert curve to map the addresses */
	hilb, err := hilbert.NewHilbert(rangeSqrt)
	if err != nil {
		slog.Error("failed to create hilbert curve map", "error", err)
		os.Exit(1)
	}

	inscan := bufio.NewScanner(cont)
	for inscan.Scan() {
		/* Ignore comments */
		if strings.HasPrefix(inscan.Text(), CommentChar) {
			continue
		}

		/* Parse the line */
		rec, err := NewScanLine(inscan.Text())
		if err != nil {
			slog.Error("failed to parse output line", "error", err)
			continue
		}

		/* Ignore addresses not contained in the target range */
		if !prefix.Contains(rec.Address) {
			continue
		}

		/* Calculate the position within the range */
		foundAs4 := rec.Address.As4()
		foundAsInt := big.NewInt(0).SetBytes(foundAs4[:])
		index := foundAsInt.Sub(foundAsInt, rangeStart).Int64()

		/* Map to an X,Y in the hilbert curve */
		x, y, err := hilb.Map(int(index))
		if err != nil {
			slog.Error("failed to map address to hilbert curve", "line", inscan.Text(), "parsed", rec, "index", index, "error", err)
			os.Exit(1)
		}

		/* Set the image color */
		if rec.Status == open {
			resultImage.Set(x, y, parseColor(*upColor))
		}
	}

	if err := inscan.Err(); err != nil {
		slog.Error("failed to scan input file", "error", err)
	}

	/* Write the PNG image to the output file */
	slog.Info("writing PNG to file...")
	png.Encode(resultFile, resultImage)
	slog.Info("done")
}
