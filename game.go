package main

import (
	"fmt"
	"math"
	"math/rand"
	"time"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
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

	gameScore *score
)

type game struct {
	world *world

	collisionBoxes *imdraw.IMDraw
}

func (g *game) init() error {
	g.world = &world{}
	g.world.init()

	gameScore = &score{
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

var drawCollisionBoxes = false

func (g *game) draw(canvas *pixelgl.Canvas) {
	// clear the canvas to black
	canvas.Clear(colornames.Black)

	g.world.draw(canvas)
	gameScore.draw()

	if win.JustPressed(pixelgl.KeyC) {
		drawCollisionBoxes = !drawCollisionBoxes
	}

	if drawCollisionBoxes {
		g.drawCollisionBoxes(canvas)
	}

	win.Clear(colornames.White)
	win.SetMatrix(pixel.IM.Scaled(pixel.ZV,
		math.Min(
			win.Bounds().W()/canvas.Bounds().W(),
			win.Bounds().H()/canvas.Bounds().H(),
		),
	).Moved(win.Bounds().Center()))
	canvas.Draw(win, pixel.IM.Moved(canvas.Bounds().Center()))

	gameScore.text.Draw(win, pixel.IM.Moved(canvas.Bounds().Min))
}

func (g *game) collisions() {
	for c := range collidables {
		checkCollisions(c)
	}
}

func (g *game) drawCollisionBoxes(t pixel.Target) {
	if g.collisionBoxes == nil {
		g.collisionBoxes = imdraw.New(nil)
	}

	g.collisionBoxes.Clear()

	for collidable := range collidables {
		g.collisionBoxes.Color = colornames.Lemonchiffon
		rect := collidable.Rect()

		g.collisionBoxes.Push(rect.Min, rect.Max)
		g.collisionBoxes.Rectangle(1)

		rect2 := sweptBroadphaseRect(collidable)

		g.collisionBoxes.Color = colornames.Indigo
		g.collisionBoxes.Push(rect2.Min, rect2.Max)
		g.collisionBoxes.Rectangle(1)
	}

	g.collisionBoxes.Draw(t)
}

func (g *game) run() {
	g.init()

	second := time.Tick(time.Second)

	canvas := pixelgl.NewCanvas(pixel.R(-1920/3, -1080/3, 1920/3, 1080/3))
	last := time.Now()
	frames := 0

	win.Canvas().SetUniform("iTime", &iTime)
	win.Canvas().SetUniform("iMouse", &iMouse)
	win.Canvas().SetUniform("iLightPos", &iLightPos)

	for i := range iMouse {
		iMouse[i] = 5
	}

	win.Canvas().SetFragmentShader(fragmentShaderLighting)

	for !win.Closed() {
		dt := time.Since(last).Seconds()
		last = time.Now()

		iTime += float32(dt)

		// lerp the camera position towards the body
		camPos = pixel.Lerp(camPos, g.world.character.body.rect.Center(), 1-math.Pow(1.0/128, dt))
		cam := pixel.IM.Moved(camPos.Scaled(-1))
		canvas.SetMatrix(cam)

		// Q: Why are these position modifiers different for each axis?
		// A: I have no clue.
		iLightPos[0] = float32(0.002 - camPos.X*0.0008)
		iLightPos[1] = float32(-0.15 - camPos.Y*0.0014)

		// slow motion with tab
		if win.Pressed(pixelgl.KeyTab) {
			dt /= 8
		}

		// restart the level on pressing enter
		if win.JustPressed(pixelgl.KeyEnter) {
			g.init()
		}

		g.collisions()
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

func (s *score) setMultiplier(multiplier int) {
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
