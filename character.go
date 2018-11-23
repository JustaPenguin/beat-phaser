package main

import (
	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
	"image/color"
)

type character struct {
	body   *body
	weapon *weapon
}

func (c *character) Vel() pixel.Vec {
	return c.body.vel
}

func (c *character) HandleCollision(x Collidable, collisionTime float64, normal pixel.Vec) {
	switch x.(type) {
	case *laser:
		c.die()
	case *wall:

		// collision Collidable and direction saved
		c.body.colliding[normal] = x

		if normal.Y == 0 {
			// collision in X. move one pixel into wall
			c.body.rect = c.body.rect.Moved(c.body.vel.ScaledXY(pixel.V(1, 0)))
		} else {
			// collision in Y. move one pixel into wall
			c.body.rect = c.body.rect.Moved(c.body.vel.ScaledXY(pixel.V(0, 1)))
		}
	}
}

func (c *character) Rect() pixel.Rect {
	return c.body.rect
}

func (c *character) die() {

}

func (c *character) init() {
	defer registerCollidable(c)

	c.body = &body{
		// phys
		gravity:   -512,
		runSpeed:  300,
		jumpSpeed: 192,
		rect:      pixel.R(-62, -74, 62, 74),
		rate:      1.0 / 10,
		dir:       1,
	}
	c.body.init()

	c.weapon = handgun
	c.weapon.init()
}

func (c *character) update(dt float64) {
	c.body.update(dt)
	c.weapon.update(dt, c.body.rect.Center(), c.body.vel)
}

func (c *character) draw(t pixel.Target) {
	c.body.draw(t)
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
