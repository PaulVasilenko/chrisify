package main

import (
	"image"
	"image/draw"
	"log"
	"os"
)

func loadImage(file string) image.Image {
	reader, err := os.Open(file)
	if err != nil {
		log.Fatalf("error loading %s: %s", file, err)
	}
	img, _, err := image.Decode(reader)
	if err != nil {
		log.Fatalf("error loading %s: %s", file, err)
	}
	return img
}

func canvasFromImage(i image.Image) *image.RGBA {
	bounds := i.Bounds()
	canvas := image.NewRGBA(bounds)
	draw.Draw(canvas, bounds, i, bounds.Min, draw.Src)

	return canvas
}
