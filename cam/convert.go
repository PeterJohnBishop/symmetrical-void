package cam

import (
	"image"
	_ "image/jpeg"
	_ "image/png"

	"github.com/qeesung/image2ascii/convert"
)

func ConvertImageToASCII(img image.Image) string {
	convertOptions := convert.DefaultOptions
	convertOptions.FixedWidth = 200
	convertOptions.FixedHeight = 80
	convertOptions.Ratio = 0.5

	converter := convert.NewImageConverter()
	return converter.Image2ASCIIString(img, &convertOptions)
}
