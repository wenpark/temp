package main

// This demonstrates a solution to resizing animated gifs.
//
// Frames in an animated gif aren't necessarily the same size, subsequent
// frames are overlayed on previous frames. Therefore, resizing the frames
// individually may cause problems due to aliasing of transparent pixels. This
// example tries to avoid this by building frames from all previous frames and
// resizing the frames as RGB.

import (
	"image"
	"image/color/palette"
	"image/draw"
	"image/gif"
	"log"
	"os"
    "fmt"
	"github.com/nfnt/resize"
)

func main() {
	process("shapes")
	//process("blob")
}

func process(filename string) {

	// Open image file.
	f, err := os.Open(filename + ".gif")
	if err != nil {
		log.Fatal(err.Error())
	}
	defer f.Close()

	// Decode the original gif.
	im, err := gif.DecodeAll(f)
	if err != nil {
		log.Fatal(err.Error())
	}

	// Create a new RGBA image to hold the incremental frames.
	firstFrame := im.Image[0].Bounds()
	b := image.Rect(0, 0, firstFrame.Dx(), firstFrame.Dy())
	img := image.NewRGBA(b)
     fmt.Println( "width:",im.Config.Width)
     fmt.Println( "height:",im.Config.Height)
     
	// Resize each frame.
	for index, frame := range im.Image {
		bounds := frame.Bounds()
		draw.Draw(img, bounds, frame, bounds.Min, draw.Over)
		im.Image[index] = ImageToPaletted(ProcessImage(720,img))
	}

	// Write resized gif.
	out, err := os.Create(filename + ".720out.fixed.gif")
	if err != nil {
		log.Fatal(err.Error())
	}
	defer out.Close()

    im.Config = image.Config{}
	gif.EncodeAll(out, im,)
}

func ProcessImage(width uint,img image.Image) image.Image {
	return resize.Resize(width, 0, img, resize.NearestNeighbor)
}

func ImageToPaletted(img image.Image) *image.Paletted {
	b := img.Bounds()
	pm := image.NewPaletted(b, palette.Plan9)
	draw.FloydSteinberg.Draw(pm, b, img, image.ZP)
	return pm
}