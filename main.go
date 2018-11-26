package main

import (
	"encoding/csv"
	"image"
	_ "image/png"
	"io"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
)

var sheetBasePath = filepath.Join("images", "sprites")

func loadAnimationSheet(name string, frameWidth float64) (sheet pixel.Picture, anims map[string][]pixel.Rect, err error) {
	// open and load the spritesheet
	sheetFile, err := os.Open(filepath.Join(sheetBasePath, name+".png"))

	if err != nil {
		return nil, nil, err
	}

	defer sheetFile.Close()

	sheetImg, _, err := image.Decode(sheetFile)

	if err != nil {
		return nil, nil, err
	}

	sheet = pixel.PictureDataFromImage(sheetImg)

	// create a slice of frames inside the spritesheet
	var frames []pixel.Rect

	for x := 0.0; x+frameWidth <= sheet.Bounds().Max.X; x += frameWidth {
		frames = append(frames, pixel.R(
			x,
			0,
			x+frameWidth,
			sheet.Bounds().H(),
		))
	}

	descFile, err := os.Open(filepath.Join(sheetBasePath, name+".csv"))

	if err != nil {
		return nil, nil, err
	}

	defer descFile.Close()

	anims = make(map[string][]pixel.Rect)

	// load the animation information, name and interval inside the spritesheet
	desc := csv.NewReader(descFile)

	for {
		anim, err := desc.Read()
		if err == io.EOF {
			break
		}

		if err != nil {
			return nil, nil, err
		}

		name := anim[0]
		start, _ := strconv.Atoi(anim[1])
		end, _ := strconv.Atoi(anim[2])

		anims[name] = frames[start : end+1]
	}

	return sheet, anims, nil
}

func randomNiceColor() pixel.RGBA {
	r := rand.Float64()
	g := rand.Float64()
	b := rand.Float64()
	l := math.Sqrt(r*r + g*g + b*b)
	if l == 0 {
		return randomNiceColor()
	}
	return pixel.RGB(r/l, g/l, b/l)
}

func main() {
	pixelgl.Run(run)
}

func run() {
	var err error

	win, err = pixelgl.NewWindow(pixelgl.WindowConfig{
		Title:  "Beat Phaser",
		Bounds: pixel.R(0, 0, 1920, 1080),
		//VSync:  true,
	})

	if err != nil {
		panic(err)
	}

	game := &game{}
	game.run()
}
