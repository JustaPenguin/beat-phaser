package main

import (
	"image/color"
	"math"
	"time"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
	"github.com/faiface/pixel/pixelgl"
)

type character struct {
	body *body
	hat  *hat
	fire *fire

	lastClick             time.Time
	increment, multiplier int

	// sounds
	pringlePhaser *soundEffect
}

func (c *character) init() {
	if c.pringlePhaser == nil {
		c.pringlePhaser = &soundEffect{
			filePath: "audio/effects/pringle-phaser.ogg",
		}
		c.pringlePhaser.load()
	}

	c.lastClick = time.Now()

	c.body = &body{
		// phys
		gravity:   -512,
		runSpeed:  128,
		jumpSpeed: 192,
		rect:      pixel.R(-62, -74, 62, 74),
		rate:      1.0 / 10,
		dir:       1,
	}
	c.body.init()

	c.hat = &hat{color: pixel.RGB(float64(255)/float64(255), 0, float64(250)/float64(255)), altColor: pixel.RGB(float64(32)/float64(255), float64(22)/float64(255), float64(249)/float64(156))}
	c.hat.init()

	c.fire = &fire{}
}

func (c *character) update(dt float64) {
	timeSinceClick := time.Since(c.lastClick).Seconds()

	if win.JustPressed(pixelgl.MouseButtonLeft) {
		c.lastClick = time.Now()

		if timeSinceClick > 60/bpm+0.05 || timeSinceClick < 60/bpm-0.05 {
			c.increment--

			if c.increment <= 0 {
				c.multiplier--
				c.increment = 8
			}

			if c.multiplier < 0 {
				c.multiplier = 0
			}

			c.fire.now(128, win.Bounds().Center().Add(c.body.rect.Center()), win.MousePosition().Add(camPos), color.White)
		} else {
			c.increment++

			if c.increment >= 8 {
				c.multiplier++
				c.increment = 0
			}

			if c.multiplier > 8 {
				c.multiplier = 8
			}

			c.fire.now(128, win.Bounds().Center().Add(c.body.rect.Center()), win.MousePosition().Add(camPos), randomNiceColor())
		}

		go c.pringlePhaser.play()
	}

	c.body.update(dt)
	c.hat.update(dt, c.body.rect.Center())
}

func (c *character) draw(t pixel.Target) {
	c.body.draw(t)
	c.hat.draw(t)
}

type hat struct {
	pos             pixel.Vec
	counter         float64
	color, altColor pixel.RGBA
}

func (h *hat) init() {

}

// @TODO hat should move up and down with animation
func (h *hat) update(dt float64, target pixel.Vec) {
	h.pos.X = target.X
	h.pos.Y = target.Y
	h.pos.Y += 6
}

func (h *hat) draw(t pixel.Target) {
	imd := imdraw.New(nil)

	imd.Color = h.color

	imd.Push(pixel.V(h.pos.X, h.pos.Y))
	imd.Push(pixel.V(h.pos.X-5, h.pos.Y+0))
	imd.Push(pixel.V(h.pos.X-5, h.pos.Y+1))
	imd.Push(pixel.V(h.pos.X-2.5, h.pos.Y+1))

	imd.Color = h.altColor

	imd.Push(pixel.V(h.pos.X-2.5, h.pos.Y+5))
	imd.Push(pixel.V(h.pos.X+2.5, h.pos.Y+5))
	imd.Push(pixel.V(h.pos.X+2.5, h.pos.Y+1))
	imd.Push(pixel.V(h.pos.X+5, h.pos.Y+1))
	imd.Push(pixel.V(h.pos.X+5, h.pos.Y+0))
	imd.Polygon(0)

	imd.Draw(t)
}

type gopherAnimationState int

const (
	idle gopherAnimationState = iota
	running
	jumping
)

type body struct {
	imd *imdraw.IMDraw

	// phys
	gravity   float64
	runSpeed  float64
	jumpSpeed float64

	rect pixel.Rect
	vel  pixel.Vec

	ground bool

	// anim
	sheet   pixel.Picture
	anims   map[string][]pixel.Rect
	rate    float64
	state   gopherAnimationState
	counter float64
	dir     float64
	frame   pixel.Rect
	sprite  *pixel.Sprite
}

func (gp *body) init() {
	if gp.sheet == nil || gp.anims == nil {
		var err error

		gp.sheet, gp.anims, err = loadAnimationSheet("spike", 124)

		if err != nil {
			panic(err)
		}
	}
}

func (gp *body) update(dt float64) {
	// control the body with keys
	ctrl := pixel.ZV
	if win.Pressed(pixelgl.KeyA) {
		ctrl.X--
	}
	if win.Pressed(pixelgl.KeyD) {
		ctrl.X++
	}
	if win.Pressed(pixelgl.KeyW) {
		ctrl.Y++
	}
	if win.Pressed(pixelgl.KeyS) {
		ctrl.Y--
	}

	// apply controls
	switch {
	case ctrl.X < 0:
		gp.vel.X = -gp.runSpeed
	case ctrl.X > 0:
		gp.vel.X = +gp.runSpeed
	default:
		gp.vel.X = 0
	}

	switch {
	case ctrl.Y < 0:
		gp.vel.Y = -gp.runSpeed
	case ctrl.Y > 0:
		gp.vel.Y = +gp.runSpeed
	default:
		gp.vel.Y = 0
	}

	// apply gravity and velocity
	gp.rect = gp.rect.Moved(gp.vel.Scaled(dt))

	// @TODO collisions with stuff looks like this, turn platforms into walls
	/*if gp.vel.Y <= 0 {
		for _, p := range platforms {
			if gp.rect.Max.X <= p.rect.Min.X || gp.rect.Min.X >= p.rect.Max.X {
				continue
			}
			if gp.rect.Min.Y > p.rect.Max.Y || gp.rect.Min.Y < p.rect.Max.Y+gp.vel.Y*dt {
				continue
			}
			gp.vel.Y = 0
			gp.rect = gp.rect.Moved(pixel.V(0, p.rect.Max.Y-gp.rect.Min.Y))
		}
	}*/

	gp.counter += dt

	// determine the new animation state
	var newState gopherAnimationState
	switch {
	case gp.vel.Len() == 0:
		newState = idle
	case gp.vel.Len() > 0:
		newState = running
	}

	// reset the time counter if the state changed
	if gp.state != newState {
		gp.state = newState
		gp.counter = 0
	}

	// determine the correct animation frame
	switch gp.state {
	case idle:
		i := int(math.Floor(gp.counter / gp.rate/2))
		gp.frame = gp.anims["Front"][i%len(gp.anims["Front"])]
	case running:
		i := int(math.Floor(gp.counter / gp.rate))
		gp.frame = gp.anims["Run"][i%len(gp.anims["Run"])]
	case jumping:
		speed := gp.vel.Y
		i := int((-speed/gp.jumpSpeed + 1) / 2 * float64(len(gp.anims["Jump"])))
		if i < 0 {
			i = 0
		}
		if i >= len(gp.anims["Front"]) {
			i = len(gp.anims["Front"]) - 1
		}
		gp.frame = gp.anims["Front"][i]
	}

	// set the facing direction of the body
	if gp.vel.X != 0 {
		if gp.vel.X > 0 {
			gp.dir = -1
		} else {
			gp.dir = +1
		}
	}
}

func (gp *body) draw(t pixel.Target) {
	if gp.imd == nil {
		gp.imd = imdraw.New(gp.sheet)
	}

	gp.imd.Clear()

	if gp.sprite == nil {
		gp.sprite = pixel.NewSprite(nil, pixel.Rect{})
	}
	// draw the correct frame with the correct position and direction
	gp.sprite.Set(gp.sheet, gp.frame)
	gp.sprite.Draw(gp.imd, pixel.IM.
		ScaledXY(pixel.ZV, pixel.V(
			gp.rect.W()/gp.sprite.Frame().W(),
			gp.rect.H()/gp.sprite.Frame().H()/1.5,
		)).
		ScaledXY(pixel.ZV, pixel.V(-gp.dir, 1)).
		Moved(gp.rect.Center()),
	)

	gp.imd.Draw(t)
}

// @TODO perhaps 'weapon'?

type laser struct {
	rect  pixel.Rect
	color color.Color

	thickness float64
}

func (l *laser) draw(imd *imdraw.IMDraw) {
	imd.Color = l.color
	imd.EndShape = imdraw.RoundEndShape

	imd.Push(pixel.V(l.rect.Min.X, l.rect.Min.Y), pixel.V(l.rect.Max.X, l.rect.Max.Y))
	imd.Line(l.thickness)

	if l.thickness > 0 {
		l.thickness = l.thickness - 0.02
	}
}

type fire struct {
	speed  float64
	origin pixel.Vec
	vector pixel.Vec

	newLaser *laser
}

func (f *fire) now(speed float64, origin pixel.Vec, vector pixel.Vec, color color.Color) {
	f.newLaser = &laser{
		color: color,
		// Minus half window size
		rect:      pixel.R(origin.X-512, origin.Y-384, vector.X-512, vector.Y-384),
		thickness: 2,
	}
}

func (f *fire) draw(imd *imdraw.IMDraw) {
	if f.newLaser != nil {
		f.newLaser.draw(imd)
	}
}
