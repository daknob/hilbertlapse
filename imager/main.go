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
	"image/color"
	"image/png"
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	"github.com/google/hilbert"
	"github.com/sirupsen/logrus"
)

func main() {
	logrus.Infof("Starting imager...")

	/* Accept the CSV file to image, default from STDIN */
	inFile := flag.String("i", "/dev/stdin", "Input CSV File")

	/* Accept the PNG file name to output as, default at STDOUT */
	outFile := flag.String("o", "/dev/stdout", "Output PNG File")
	flag.Parse()

	/* Open the CSV file */
	logrus.Infof("Loading file from %s...", *inFile)
	cont, err := ioutil.ReadFile(*inFile)
	if err != nil {
		logrus.Fatalf("Could not read from %s...: %v", *inFile, err)
	}
	logrus.Infof("File loaded!")

	/* Create a new PNG image in memory */
	logrus.Infof("Creating image file...")
	resultImage := image.NewRGBA(
		image.Rectangle{
			image.Point{0, 0},
			image.Point{255, 255},
		},
	)

	/* Create the output file, and open it for writting */
	resultFile, err := os.Create(*outFile)
	if err != nil {
		logrus.Fatalf(
			"Failed to create image file (%s) for writting output: %v",
			*outFile, err,
		)
	}

	/* Create a new Hilbert map, 256x256 to map the addresses in */
	hilb, err := hilbert.NewHilbert(256)
	if err != nil {
		logrus.Fatalf("Failed to create Hilbert Map: %v", err)
	}

	/* Split the CSV file, line by line, into a slice */
	lines := strings.Split(string(cont), "\n")

	/* Remove the last element (empty) */
	lines = lines[:len(lines)-1]

	/* For every line in the CSV file, */
	for _, l := range lines {

		/* Split the fields based on the "," character */
		fields := strings.Split(l, ",")

		/* Extract the coordinates (/24, /32) */
		x := strings.Split(fields[0], ".")[2]
		y := strings.Split(fields[0], ".")[3]

		/* Convert them to integers */
		xint, err := strconv.Atoi(x)
		if err != nil {
			/*
			 * If there's an error, exit. Feel free to modify code accordingly
			 * to instead skip the line, using continue.
			 */
			logrus.Fatalf("Failed to convert string to int: %v", err)
		}
		yint, err := strconv.Atoi(y)
		if err != nil {
			logrus.Fatalf("Failed to convert string to int: %v", err)
		}

		/*
		 * Get the X, Y coordinates of the final PNG, by mapping the X and Y
		 * using the Hilbert map created earlier
		 */
		actx, acty, err := hilb.Map(xint*256 + yint)
		if err != nil {
			/* Exit on error, feel free to modify code as before */
			logrus.Fatalf("Failed to hilbert-map point: %v", err)
		}

		/* Check the CSV field and determine if the pixel should be "on" */
		if fields[1] == "up" {
			/* If yes, color it green */
			resultImage.Set(actx, acty, color.RGBA{50, 200, 50, 255})
		} else {
			/*
			 * Otherwise, color it dark gray. It is not being colored as black,
			 * because there will be no difference between missing elements
			 * from the CSV, and pixels that are marked explicitly as "off".
			 */
			resultImage.Set(actx, acty, color.RGBA{50, 50, 50, 255})
		}

	}

	/* Write the PNG image to the output file */
	logrus.Infof("Writting PNG to file...")
	png.Encode(resultFile, resultImage)
	logrus.Infof("Done!")
}
