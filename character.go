package main

import (
	"image/color"
	"math"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
	"github.com/faiface/pixel/pixelgl"
	"golang.org/x/image/colornames"
)

type character struct {
	body   *body
	hat    *hat
	weapon *weapon
}

func (c *character) HandleCollision(x Collidable) {
	switch a := x.(type) {
	case *laser:
		c.die()
	case *platform:

		r := x.Rect()

		if int(c.body.rect.Center().Y) >= int(r.Max.Y) {
			// above
			c.body.rect = c.body.rect.Moved(pixel.V(0, a.rect.Max.Y-c.body.rect.Min.Y))
			c.body.vel.Y = 0
		} else if int(c.body.rect.Center().Y) <= int(r.Min.Y) {
			// below
			c.body.rect = c.body.rect.Moved(pixel.V(0, a.rect.Min.Y-c.body.rect.Max.Y-5)) // -5 - hat
			c.body.vel.Y = 0
		} else if int(c.body.rect.Center().X) <= int(r.Min.X) {
			// left
			c.body.rect = c.body.rect.Moved(pixel.V(a.rect.Min.X-c.body.rect.Max.X, 0))
			c.body.vel.X = 0
		} else {
			// right
			c.body.rect = c.body.rect.Moved(pixel.V(a.rect.Max.X-c.body.rect.Min.X, 0))
			c.body.vel.X = 0
		}
	}
}

func (c *character) Rect() pixel.Rect {
	r := c.body.rect
	r.Max = r.Max.Add(pixel.V(0, 5)) // hat

	return r
}

func (c *character) die() {
	c.hat.color = colornames.Red
	c.hat.altColor = colornames.Black
}

func (c *character) init() {
	defer registerCollidable(c)

	c.body = &body{
		// phys
		gravity:   -512,
		runSpeed:  64,
		jumpSpeed: 192,
		rect:      pixel.R(-6, -7, 6, 7),
		rate:      1.0 / 10,
		dir:       +1,
	}
	c.body.init()

	c.hat = &hat{color: pixel.RGB(float64(255)/float64(255), 0, float64(250)/float64(255)), altColor: pixel.RGB(float64(32)/float64(255), float64(22)/float64(255), float64(249)/float64(156))}
	c.hat.init()

	c.weapon = handgun
	c.weapon.init()
}

func (c *character) update(dt float64) {
	c.body.update(dt)
	c.hat.update(dt, c.body.rect.Center())
	c.weapon.update(dt, c.body.rect.Center(), c.body.vel)
}

func (c *character) draw(t pixel.Target) {
	c.body.draw(t)
	c.hat.draw(t)
	c.weapon.draw(t)
}

type hat struct {
	pos             pixel.Vec
	counter         float64
	color, altColor color.Color
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

		gp.sheet, gp.anims, err = loadAnimationSheet("gopher", 12)

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

	gp.counter += dt

	// determine the new animation state
	var newState gopherAnimationState
	switch {
	case !gp.ground:
		newState = jumping
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
		gp.frame = gp.anims["Front"][0]
	case running:
		i := int(math.Floor(gp.counter / gp.rate))
		gp.frame = gp.anims["Run"][i%len(gp.anims["Run"])]
	case jumping:
		speed := gp.vel.Y
		i := int((-speed/gp.jumpSpeed + 1) / 2 * float64(len(gp.anims["Jump"])))
		if i < 0 {
			i = 0
		}
		if i >= len(gp.anims["Jump"]) {
			i = len(gp.anims["Jump"]) - 1
		}
		gp.frame = gp.anims["Jump"][i]
	}

	// set the facing direction of the body
	if gp.vel.X != 0 {
		if gp.vel.X > 0 {
			gp.dir = +1
		} else {
			gp.dir = -1
		}
	}
}

func (gp *body) draw(t pixel.Target) {
	x := imdraw.New(gp.sheet)

	if gp.sprite == nil {
		gp.sprite = pixel.NewSprite(nil, pixel.Rect{})
	}
	// draw the correct frame with the correct position and direction
	gp.sprite.Set(gp.sheet, gp.frame)
	gp.sprite.Draw(x, pixel.IM.
		ScaledXY(pixel.ZV, pixel.V(
			gp.rect.W()/gp.sprite.Frame().W(),
			gp.rect.H()/gp.sprite.Frame().H(),
		)).
		ScaledXY(pixel.ZV, pixel.V(-gp.dir, 1)).
		Moved(gp.rect.Center()),
	)
	x.Draw(t)
}

func todegrees(rads float64) float64 {
	return rads * (180 / math.Pi)
}

func angleBetweenVectors(v1, v2 pixel.Vec) float64 {
	angle := math.Atan2(v2.Y, v2.X) - math.Atan2(v1.Y, v2.X)

	if angle < 0 {
		angle += 2 * math.Pi
	}

	return angle
}

func getMouseAngleFromCenter() float64 {
	a := angleBetweenVectors(pixel.V(0, 0), win.Bounds().Center().Sub(win.MousePosition()))

	if win.MousePosition().X < win.Bounds().Center().X {
		a -= math.Pi
	}

	return a
}
