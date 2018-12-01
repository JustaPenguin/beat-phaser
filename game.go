package main

import (
	"fmt"
	"math"
	"math/rand"
	"time"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
	"github.com/faiface/pixel/pixelgl"
	"golang.org/x/image/colornames"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

var (
	win    *pixelgl.Window
	camPos pixel.Vec

	playerScore    *score
	playerSpawnPos = pixel.V(-625, -50)
)

type game struct {
	world *world

	collisionBoxes *imdraw.IMDraw
}

func (g *game) init() error {
	g.world = &world{}
	g.world.init()
	ded = false

	playerScore = &score{}
	playerScore.init()

	// camera
	camPos = playerSpawnPos

	return nil
}

func (g *game) update(dt float64) {
	g.world.update(dt)
	playerScore.update(dt)
}

var drawCollisionBoxes = false

func (g *game) draw(canvas *pixelgl.Canvas) {
	// clear the canvas to black
	canvas.Clear(colornames.Black)

	g.world.draw(canvas)

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
	playerScore.draw(win, canvas)
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

	/*win.Canvas().SetUniform("iTime", &iTime)

	win.Canvas().SetUniform("iMouse", &iMouse)
	win.Canvas().SetUniform("iLightPos", &iLightPos)

	for i := range iMouse {
		iMouse[i] = 5
	}

	win.Canvas().SetFragmentShader(fragmentShaderLighting)
	*/

	frameLimit := time.Tick(time.Second / 144)

	for !win.Closed() {
		dt := time.Since(last).Seconds()
		last = time.Now()

		//iTime += float32(dt)

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
			g.destroy()
			g.init()
		}

		g.collisions()
		g.update(dt)

		g.draw(canvas)
		win.Update()

		frames++
		select {
		case <-second:
			win.SetTitle(fmt.Sprintf("Beat Phaser - FPS: %d", frames))
			frames = 0
		default:
		}

		<-frameLimit
	}
}

func (g *game) destroy() {
	g.world.destroy()
}
