package main

import (
	"math"

	"github.com/faiface/pixel"
)

type Collidable interface {
	Rect() pixel.Rect
	Vel() pixel.Vec
	HandleCollision(obj Collidable, collisionTime float64, normal pixel.Vec)
}

var collidables = make(map[Collidable]bool)

func deregisterCollidable(c Collidable) {
	delete(collidables, c)
}

func registerCollidable(c Collidable) {
	collidables[c] = true
}

func checkCollisions(c Collidable) {
	for x := range collidables {
		bpb := sweptBroadphaseRect(c)

		if aabbCheck(bpb, x.Rect()) {
			collisionTime, normal := sweptAABB(c, x)

			if collisionTime < 1 {
				x.HandleCollision(c, collisionTime, normal)
				c.HandleCollision(x, collisionTime, normal)
			}
		}
	}
}

// aabbCheck is a simple collision test for two axis aligned bounding boxes. it may report false positives.
func aabbCheck(b1, b2 pixel.Rect) bool {
	b1 = b1.Norm()
	b2 = b2.Norm()

	return b1.Intersect(b2).Area() > 0
}

// sweptBroadphaseRect calculates the broadphase area for a collidable. This creates a collision box which has been scaled
// up to include where the box has travelled under its velocity.
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

// sweptAABB checks for a collision between a moving AABB and a static AABB,
// returning:
// - the collisionTime: a value between 0 and 1 means a collision occurred
// - the normal of the collided surface
//
// ref: https://www.gamedev.net/articles/programming/general-and-gameplay-programming/swept-aabb-collision-detection-and-response-r3084/
func sweptAABB(moving, static Collidable) (collisionTime float64, normal pixel.Vec) {
	movingRect := moving.Rect().Norm()
	staticRect := static.Rect().Norm()

	var entryInv, exitInv pixel.Vec

	if moving.Vel().X > 0 {
		entryInv.X = staticRect.Min.X - (movingRect.Min.X + movingRect.W())
		exitInv.X = (staticRect.Min.X + staticRect.W()) - movingRect.Min.X
	} else {
		entryInv.X = (staticRect.Min.X + staticRect.W()) - movingRect.Min.X
		exitInv.X = staticRect.Min.X - (movingRect.Min.X + movingRect.W())
	}

	if moving.Vel().Y > 0 {
		entryInv.Y = staticRect.Min.Y - (movingRect.Min.Y + movingRect.H())
		exitInv.Y = (staticRect.Min.Y + staticRect.H()) - movingRect.Min.Y
	} else {
		entryInv.Y = (staticRect.Min.Y + staticRect.H()) - movingRect.Min.Y
		exitInv.Y = staticRect.Min.Y - (movingRect.Min.Y + movingRect.H())
	}

	var entry, exit pixel.Vec

	if moving.Vel().X == 0 {
		entry.X = math.Inf(-1)
		exit.X = math.Inf(1)
	} else {
		entry.X = entryInv.X / moving.Vel().X
		exit.X = exitInv.X / moving.Vel().X
	}

	if moving.Vel().Y == 0 {
		entry.Y = math.Inf(-1)
		exit.Y = math.Inf(1)
	} else {
		entry.Y = entryInv.Y / moving.Vel().Y
		exit.Y = exitInv.Y / moving.Vel().Y
	}

	// earliest time of collision
	entryTime := math.Max(entry.X, entry.Y)
	exitTime := math.Min(exit.X, exit.Y)

	if (movingRect.Intersect(staticRect).Area() == 0) && (entryTime > exitTime || entry.X < 0 && entry.Y < 0 || entry.X > 1 || entry.Y > 1) {
		// no collision
		return 1.0, pixel.ZV
	} else {
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

		return entryTime, normal
	}
}
