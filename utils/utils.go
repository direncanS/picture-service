package utils

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"

	"github.com/nfnt/resize"
)

func CompressImage(img image.Image) ([]byte, error) {
	crrWidth := img.Bounds().Dx()
	crrHeihgt := img.Bounds().Dy()

	newWidth := crrWidth / 2
	newHeight := crrHeihgt / 2

	resizedImg := resize.Resize(uint(newWidth), uint(newHeight), img, resize.Lanczos2)

	var resizedImageBuffer bytes.Buffer

	err := jpeg.Encode(&resizedImageBuffer, resizedImg, nil)
	if err != nil {
		fmt.Println("Error encoding resized image:", err)
		return nil, err
	}

	return resizedImageBuffer.Bytes(), nil
}
