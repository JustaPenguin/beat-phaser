package main

import (
	"fmt"
	"github.com/faiface/pixel"
	"math"
)

type Collidable interface {
	Rect() pixel.Rect
	Vel() pixel.Vec
	HandleCollision(obj Collidable, collisionTime float64, normal pixel.Vec)
}

var collidables = make(map[Collidable]pixel.Rect)

func deregisterCollidable(c Collidable) {
	delete(collidables, c)
}

func registerCollidable(c Collidable) {
	collidables[c] = c.Rect()
}

func checkCollisions(c Collidable) {
	for x := range collidables {
		bpb := sweptBroadphaseRect(c)

		if aabbCheck(bpb, x.Rect()) {
			normal, collision := sweptAABB(c, x)

			//fmt.Printf("(%T) -> (%T) SWEEP\n", x, c)

			if collision < 1 {
				fmt.Printf("(%T) -> (%T) collision at %s\n", x, c, normal)
				x.HandleCollision(c, collision, normal)
				c.HandleCollision(x, collision, normal)
			}
		}
	}
}

func aabbCheck(b1, b2 pixel.Rect) bool {
	b1 = b1.Norm()
	b2 = b2.Norm()

	return !(b1.Min.X+b1.W() < b2.Min.X || b1.Min.X > b2.Min.X+b2.W() || b1.Min.Y+b1.H() < b2.Min.Y || b1.Min.Y > b2.Min.Y+b2.H())
}

func sweptBroadphaseRect(b Collidable) pixel.Rect {
	r := pixel.Rect{}
	bRect := b.Rect().Norm()

	if b.Vel().X > 0 {
		r.Min.X = bRect.Min.X
		r.Max.X = bRect.Max.X + b.Vel().X
	} else {
		r.Min.X = bRect.Min.X + b.Vel().X
		r.Max.X = bRect.Max.X - b.Vel().X
	}

	if b.Vel().Y > 0 {
		r.Min.Y = bRect.Min.Y
		r.Max.Y = bRect.Max.Y + b.Vel().Y
	} else {
		r.Min.Y = bRect.Min.Y + b.Vel().Y
		r.Max.Y = bRect.Max.Y - b.Vel().Y
	}

	return r
}

func sweptAABB(b1, b2 Collidable) (normal pixel.Vec, collision float64) {
	b1Rect := b1.Rect().Norm()
	b2Rect := b2.Rect().Norm()

	var entryInv, exitInv pixel.Vec

	if b1.Vel().X > 0 {
		entryInv.X = b2Rect.Min.X - (b1Rect.Min.X + b1Rect.W())
		exitInv.X = (b2Rect.Min.X + b2Rect.W()) - b1Rect.Min.X
	} else {
		entryInv.X = (b2Rect.Min.X + b2Rect.W()) - b1Rect.Min.X
		exitInv.X = b2Rect.Min.X - (b1Rect.Min.X + b1Rect.W())
	}

	if b1.Vel().Y > 0 {
		entryInv.Y = b2Rect.Min.Y - (b1Rect.Min.Y + b1Rect.H())
		exitInv.Y = (b2Rect.Min.Y + b2Rect.H()) - b1Rect.Min.Y
	} else {
		entryInv.Y = (b2Rect.Min.Y + b2Rect.H()) - b1Rect.Min.Y
		exitInv.Y = b2Rect.Min.Y - (b1Rect.Min.Y + b1Rect.H())
	}

	var entry, exit pixel.Vec

	if b1.Vel().X == 0 {
		entry.X = math.Inf(-1)
		exit.X = math.Inf(1)
	} else {
		entry.X = entryInv.X / b1.Vel().X
		exit.X = exitInv.X / b1.Vel().X
	}

	if b1.Vel().Y == 0 {
		entry.Y = math.Inf(-1)
		exit.Y = math.Inf(1)
	} else {
		entry.Y = entryInv.Y / b1.Vel().Y
		exit.Y = exitInv.Y / b1.Vel().Y
	}

	// earliest time of collision
	entryTime := math.Max(entry.X, entry.Y)
	exitTime := math.Min(exit.X, exit.Y)

	// no collision
	if entryTime > exitTime || entry.X < 0 && entry.Y < 0 || entry.X > 1 || entry.Y > 1 {
		return pixel.ZV, 1.0
	} else {
		println("prettysure theres a collision")
		// collision
		// calculate normal
		if entry.X > entry.Y {
			if entryInv.X < 0 {
				normal.X = 1
				normal.Y = 0
			} else {
				normal.X = -1
				normal.Y = 0
			}
		} else {
			if entryInv.Y < 0 {
				normal.X = 0
				normal.Y = 1
			} else {
				normal.X = 0
				normal.Y = -1
			}
		}

		return normal, entryTime
	}
}

func updatePositions() {
	for x := range collidables {
		collidables[x] = x.Rect()
	}
}

func didCollide(a pixel.Rect, b pixel.Rect) bool {
	return a.Norm().Intersect(b.Norm()).Area() != 0
}
