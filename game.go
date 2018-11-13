package main

import (
	"fmt"
	"golang.org/x/image/colornames"
	"math"
	"math/rand"
	"time"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
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
	var err error

	win, err = pixelgl.NewWindow(pixelgl.WindowConfig{
		Title:  "Platformer",
		Bounds: pixel.R(0, 0, 1024, 768),
		VSync:  true,
	})

	if err != nil {
		return err
	}

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

func (g *game) draw() {

}

func (g *game) run() {
	g.init()

	canvas := pixelgl.NewCanvas(pixel.R(-160/2, -160/2, 160/2, 160/2))
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
