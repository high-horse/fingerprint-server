package main

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"io"

	"github.com/jtejido/sourceafis"
)

func WriteAndDecode(data []byte) {
	
}

func LoadImageFromBytes(data []byte) (*sourceafis.Image, error) {
	reader := bytes.NewReader(data)

	// Try JPEG first
	if img, err := jpeg.Decode(reader); err == nil {
		return convertToSourceAFISImage(img)
	}

	// Reset reader and try PNG
	reader.Seek(0, io.SeekStart)
	if img, err := png.Decode(reader); err == nil {
		return convertToSourceAFISImage(img)
	}

	return nil, fmt.Errorf("unsupported image format - must be JPEG or PNG")
}

func convertToSourceAFISImage(img image.Image) (*sourceafis.Image, error) {
	bounds := img.Bounds()
	gray := image.NewGray(bounds)
	for x := 0; x < bounds.Max.X; x++ {
		for y := 0; y < bounds.Max.Y; y++ {
			c := img.At(x, y)
			gray.Set(x, y, color.GrayModel.Convert(c))
		}
	}
	return sourceafis.NewFromGray(gray)
}