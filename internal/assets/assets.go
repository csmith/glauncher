package assets

import (
	"bytes"
	"embed"
	"fmt"
	"image"
	"image/color"
	"image/draw"

	"github.com/srwiley/oksvg"
	"github.com/srwiley/rasterx"
)

//go:embed *.svg
var svgs embed.FS

var cache map[string]image.Image

func init() {
	cache = make(map[string]image.Image)
}

func load(name string, size int) image.Image {
	key := fmt.Sprintf("%s:%d", name, size)
	if img, ok := cache[key]; ok {
		return img
	}

	data, err := svgs.ReadFile(name + ".svg")
	if err != nil {
		return solidIcon(size, color.NRGBA{R: 80, G: 80, B: 100, A: 255})
	}

	icon, err := oksvg.ReadIconStream(bytes.NewReader(data))
	if err != nil {
		return solidIcon(size, color.NRGBA{R: 80, G: 80, B: 100, A: 255})
	}

	icon.SetTarget(0, 0, float64(size), float64(size))
	img := image.NewRGBA(image.Rect(0, 0, size, size))
	scanner := rasterx.NewScannerGV(size, size, img, img.Bounds())
	dasher := rasterx.NewDasher(size, size, scanner)
	icon.Draw(dasher, 1)

	cache[key] = img
	return img
}

func solidIcon(size int, c color.NRGBA) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, size, size))
	draw.Draw(img, img.Bounds(), &image.Uniform{C: c}, image.Point{}, draw.Src)
	return img
}

func Arch(size int) image.Image        { return load("arch", size) }
func ArchAlt(size int) image.Image     { return load("arch-alt", size) }
func Calculator(size int) image.Image  { return load("calculator", size) }
func Code(size int) image.Image        { return load("code", size) }
func Error(size int) image.Image       { return load("error", size) }
func Folder(size int) image.Image      { return load("folder", size) }
func Loading(size int) image.Image     { return load("loading", size) }
func Placeholder(size int) image.Image { return load("placeholder", size) }
func Copy(size int) image.Image        { return load("copy", size) }
func Search(size int) image.Image      { return load("search", size) }
