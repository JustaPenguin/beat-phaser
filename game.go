package main

import (
	"fmt"
	"math"
	"math/rand"
	"time"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
	"github.com/faiface/pixel/text"
	"golang.org/x/image/colornames"
	"golang.org/x/image/font/basicfont"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

var (
	win    *pixelgl.Window
	camPos pixel.Vec
	bpm    float64 = 103
)

type game struct {
	score *score

	world *world
}

func (g *game) init() error {
	g.world = &world{}
	g.world.init()

	g.score = &score{
		multiplier: 0,
		pos:        pixel.V(0, 0),
	}

	// camera
	camPos = pixel.ZV

	return nil
}

func (g *game) update(dt float64) {
	g.world.update(dt)
}

func (g *game) draw(canvas *pixelgl.Canvas) {
	// clear the canvas to black
	canvas.Clear(colornames.Black)

	// draw to imd
	g.world.draw(canvas)
	g.score.draw()

	win.Clear(colornames.White)
	win.SetMatrix(pixel.IM.Scaled(pixel.ZV,
		math.Min(
			win.Bounds().W()/canvas.Bounds().W(),
			win.Bounds().H()/canvas.Bounds().H(),
		),
	).Moved(win.Bounds().Center()))
	canvas.Draw(win, pixel.IM.Moved(canvas.Bounds().Center()))

	g.score.text.Draw(win, pixel.IM.Moved(canvas.Bounds().Min))
}

func (g *game) run() {
	g.init()

	canvas := pixelgl.NewCanvas(pixel.R(-1920/2, -1080/2, 1920/2, 1080/2))
	last := time.Now()
	frames := 0
	second := time.Tick(time.Second)

	for !win.Closed() {
		dt := time.Since(last).Seconds()
		last = time.Now()

		// lerp the camera position towards the body
		camPos = pixel.Lerp(camPos, g.world.character.body.rect.Center(), 1-math.Pow(1.0/128, dt))
		cam := pixel.IM.Moved(camPos.Scaled(-1))
		canvas.SetMatrix(cam)

		// slow motion with tab
		if win.Pressed(pixelgl.KeyTab) {
			dt /= 8
		}

		// restart the level on pressing enter
		if win.JustPressed(pixelgl.KeyEnter) {
			g.init()
		}

		g.update(dt)
		g.draw(canvas)

		win.Update()

		frames++
		select {
		case <-second:
			win.SetTitle(fmt.Sprintf("FPS: %d", frames))
			frames = 0
		default:
		}
	}
}

type score struct {
	multiplier int
	pos        pixel.Vec

	text *text.Text
}

func (s *score) update(multiplier int) {
	s.multiplier = multiplier
}

func (s *score) draw() {
	atlas := text.NewAtlas(
		basicfont.Face7x13,
		[]rune{'0', '1', '2', '3', '4', '5', '6', '7', '8', 'x'},
	)

	s.text = text.New(s.pos, atlas)

	// @TODO colours for scores, perhaps a little animated multi tone stuff for big ones
	switch s.multiplier {
	case 1:
		s.text.Color = colornames.Aqua
	case 2:
		s.text.Color = colornames.Coral

	}

	fmt.Fprintf(s.text, "%dx", s.multiplier)
}
