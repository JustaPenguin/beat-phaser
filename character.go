package main

import (
	"image/color"
	"time"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
)

var characterIsOutside = false

type character struct {
	body   *body
	weapon *weapon

	tick <-chan time.Time
}

func (c *character) Vel() pixel.Vec {
	return c.body.vel
}

func (c *character) HandleCollision(x Collidable, collisionTime float64, normal pixel.Vec) {
	switch collidable := x.(type) {
	case *wall:
		if normal.Y == 0 {
			// collision in X. move back by c.body.vel (with a negated x)
			c.body.rect = c.body.rect.Moved(c.body.vel.ScaledXY(pixel.V(-1, 0)))
		} else {
			// collision in Y. move back by c.body.vel (with a negated y)
			c.body.rect = c.body.rect.Moved(c.body.vel.ScaledXY(pixel.V(0, -1)))
		}
	case *enemy:
		// damage

		select {
		case <-c.tick:
			if !collidable.isAttacking {
				return
			}

			if playerScore.multiplier > 1 {
				playerScore.setMultiplier(playerScore.multiplier - 1)
			} else {
				c.body.health -= 20

				playerScore.incrementScore(-20.0)

				if c.body.health <= 0 {
					c.die()
				}
			}
		default:

		}
	}
}

func (c *character) Rect() pixel.Rect {
	return c.body.rect
}

func (c *character) die() {
	ded = true
	isDodging = false
	playerScore.changeTrack(nightOnTheDocksAudio)
}

func (c *character) init() {
	defer registerCollidable(c)

	c.body = &body{
		// phys
		gravity:   -512,
		runSpeed:  300,
		jumpSpeed: 192,
		rect:      pixel.R(-62, -74, 62, 74).Moved(playerSpawnPos),
		rate:      1.0 / 10,
		dir:       -1,
	}
	c.body.init()

	c.weapon = handgun
	c.weapon.init()

	c.tick = time.Tick(time.Millisecond * 200)
}

func (c *character) update(dt float64) {
	c.body.update(dt)
	c.weapon.update(dt, c.body.shootPos, c.body.vel)

	characterIsOutside = c.body.rect.Norm().Intersect(streetBoundingRect.Norm()).Area() > 0
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
