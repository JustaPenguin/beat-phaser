package main

import (
	"github.com/faiface/pixel"
)

type Collidable interface {
	Rect() pixel.Rect
	HandleCollision(Collidable)
}

var collidables = make(map[Collidable]pixel.Rect)

func deregisterCollidable(c Collidable) {
	delete(collidables, c)
}

func registerCollidable(c Collidable) {
	collidables[c] = c.Rect()
}

func collision(c Collidable) []Collidable {
	var cs []Collidable

	for x, r := range collidables {
		if x == c {
			continue
		}

		if didCollide(c.Rect().Norm().Union(collidables[c].Norm()), x.Rect().Norm().Union(r.Norm())) {
			cs = append(cs, x)
		}
	}

	return cs
}

func updatePositions() {
	for x := range collidables {
		collidables[x] = x.Rect()
	}
}

func didCollide(a pixel.Rect, b pixel.Rect) bool {
	return a.Norm().Intersect(b.Norm()).Area() != 0
}
