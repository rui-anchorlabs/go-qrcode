// go-qrcode
// Copyright 2014 Tom Harwood

package main

import (
	"bytes"
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"

	qrcode "github.com/skip2/go-qrcode"
)

func main() {
	// content := flag.String("c", "", "content to encode")
	outFile := flag.String("o", "", "out PNG file prefix, empty for stdout")
	size := flag.Int("s", 256, "image size (pixel)")
	textArt := flag.Bool("t", false, "print as text-art on stdout")
	negative := flag.Bool("i", false, "invert black and white")
	disableBorder := flag.Bool("d", false, "disable QR Code border")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, `qrcode -- QR Code encoder in Go
Forked from https://github.com/skip2/go-qrcode
Modified by Anchorage.

Flags:
`)
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, `
Usage:
  1. Arguments except for flags are joined by " " and used to generate QR code.
     Default output is STDOUT, pipe to imagemagick command "display" to display
     on any X server.

       qrcode hello word | display

  2. Save to file if "display" not available:

       qrcode "homepage: https://github.com/skip2/go-qrcode" > out.png

`)
	}
	flag.Parse()

	if len(flag.Args()) == 0 {
		flag.Usage()
		checkError(fmt.Errorf("Error: no content given"))
	}

	args := flag.Args()
	content := args[0]
	logoPath := args[1]

	var err error
	var code *qrcode.QRCode
	code, err = qrcode.New(content, qrcode.Highest)
	checkError(err)

	if *disableBorder {
		code.DisableBorder = true
	}

	if *textArt {
		art := code.ToString(*negative)
		fmt.Println(art)
		return
	}

	if *negative {
		code.ForegroundColor, code.BackgroundColor = code.BackgroundColor, code.ForegroundColor
	}

	var buf bytes.Buffer
	img := code.Image(*size)

	file, err := os.Open(logoPath)
	checkError(err)
	defer file.Close()

	logo, _, err := image.Decode(file)
	checkError(err)
	overlayLogo(img, logo, code)

	err = png.Encode(&buf, img)
	checkError(err)

	if *outFile == "" {
		os.Stdout.Write(buf.Bytes())
	} else {
		var fh *os.File
		fh, err = os.Create(*outFile + ".png")
		checkError(err)
		defer fh.Close()
		fh.Write(buf.Bytes())
	}
}

func checkError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}

var GreyThreshold = 30
var AlphaThreshold = 2000

// overlayLogo blends logo to the center of the QR code.
func overlayLogo(dst, src image.Image, qrcode *qrcode.QRCode) {
	grey := uint32(^uint16(0)) * uint32(GreyThreshold) / 100
	alphaOffset := uint32(AlphaThreshold)
	offset := dst.Bounds().Max.X/2 - src.Bounds().Max.X/2
	for x := 0; x < src.Bounds().Max.X; x++ {
		for y := 0; y < src.Bounds().Max.Y; y++ {
			if r, g, b, alpha := src.At(x, y).RGBA(); alpha > alphaOffset {
				col := qrcode.ForegroundColor
				if r > grey && g > grey && b > grey {
					col = color.White
				}
				dst.(*image.Paletted).Set(x+offset, y+offset, col)
			}
		}
	}
}
