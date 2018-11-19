package main

import (
	"github.com/faiface/pixel"
)

type Collidable interface {
	Rect() pixel.Rect
	HandleCollision(Collidable)
//	Dirty() bool
}

var collidables = make(map[Collidable]pixel.Rect)

func DeregisterCollidable(c Collidable) {
	delete(collidables, c)
}

func RegisterCollidable(c Collidable) {
	collidables[c] = c.Rect()
}

func Collision(c Collidable) []Collidable {
//	if !c.Dirty() {
//		return nil
//	}

	var cs []Collidable

	for x, r := range collidables {
		if x == c/* || !x.Dirty() */{
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

