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
	"flag"
	"image"
	"image/color/palette"
	"image/draw"
	"image/gif"
	"image/png"
	"log/slog"
	"os"
	"strings"
)

func main() {
	slog.Info("starting animator...")

	/* The file that contains a list of all PNGs to be animated */
	srcList := flag.String("s", "/dev/stdin", "File to read list of PNG names")

	/* The name of the output animated GIF file */
	outFile := flag.String("o", "animated.gif", "Output GIF name")
	flag.Parse()

	/* Read the file with the list of all PNG names */
	slog.Info("reading input file list", "file", *srcList)
	fl, err := os.ReadFile(*srcList)
	if err != nil {
		slog.Error("failed to read file list", "file", *srcList, "error", err)
		os.Exit(1)
	}

	/* Split the file names on the new line */
	files := strings.Split(string(fl), "\n")
	/* Remove the last element (empty) */
	files = files[:len(files)-1]

	/* Print the amount of loaded files to animate */
	slog.Info("loaded files", "count", len(files))

	/* Create a GIF image in memory */
	outputGIF := &gif.GIF{}

	/* For every file name, */
	for _, fn := range files {
		/* open it, */
		inPNG, err := os.Open(fn)
		if err != nil {
			/* We ignore all files that cannot be opened */
			slog.Warn("could not open file", "file", fn, "error", err)
			continue
		}
		/* decode the PNG in it, */
		inIMG, err := png.Decode(inPNG)
		if err != nil {
			/* We ignore all non-PNG files */
			slog.Warn("could not read PNG file", "file", fn, "error", err)
			continue
		}

		/* Convert the PNG to GIF */
		slog.Info("converting file to GIF", "file", fn)
		pIMG := image.NewPaletted(inIMG.Bounds(), palette.Plan9)
		draw.Draw(pIMG, pIMG.Rect, inIMG, inIMG.Bounds().Min, draw.Over)

		/* Add the new GIF frame to the GIF file, in-memory */
		outputGIF.Image = append(outputGIF.Image, pIMG)

		/* Add a delay of 0 between the frames */
		outputGIF.Delay = append(outputGIF.Delay, 0)
	}

	slog.Info("encoding complete")
	slog.Info("writing GIF...")

	/* Open the GIF file for writting, create if necessary */
	out, err := os.OpenFile(*outFile, os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		slog.Error("failed to open file for writing output", "file", *outFile, "error", err)
		os.Exit(1)
	}

	/* Encode the GIF, and write it to file */
	err = gif.EncodeAll(out, outputGIF)
	if err != nil {
		slog.Error("failed to encode output GIF", "error", err)
		os.Exit(1)
	}

	/* Close the output GIF file */
	out.Close()

	slog.Info("done")
}
