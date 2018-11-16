package main

import (
	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
	rand2 "math/rand"
)

type enemiesCollection struct {
	enemies []enemy

	counter float64
	step 	float64
}

func(e *enemiesCollection) init() {
	e.enemies = append(e.enemies, enemy{
		initialPos: pixel.Vec{X: 0, Y: 0},
		pos: pixel.Vec{X: 0, Y: 0},
		moveSpeed: 0.2,

		color1: randomNiceColor(),
		color2: randomNiceColor(),


		imd: imdraw.New(nil),
	})

	e.step = 20

	for i :=range e.enemies {
		e.enemies[i].init()
	}
}

func(e *enemiesCollection) update(dt float64, targetPos pixel.Vec) {
	e.counter += dt

	// e.step seconds have passed, add a new enemy (and increase spawn rate)
	// max enemies 50 (could be more but hey)
	if len(e.enemies) <= 50 {
		if e.counter > e.step {
			e.enemies = append(e.enemies, enemy{
				initialPos: pixel.Vec{X: 0, Y: 0},
				pos:       pixel.Vec{X: 0, Y: 0},
				moveSpeed: 0.2,

				color1: randomNiceColor(),
				color2: randomNiceColor(),

				imd: imdraw.New(nil),
			})

			e.step -= 1

			if e.step <= 2 {
				e.step = 2
			}

			e.counter = 0
		}
	}

	for i := range e.enemies {
		e.enemies[i].update(dt, targetPos)
	}
}

func(e *enemiesCollection) draw(t pixel.Target) {
	for i := range e.enemies {
		e.enemies[i].draw(t)
	}
}

type enemy struct {
	pos pixel.Vec
	moveSpeed float64

	initialPos pixel.Vec
	midPoint pixel.Vec
	target pixel.Vec

	counter float64
	step float64

	color1 pixel.RGBA
	color2 pixel.RGBA

	imd *imdraw.IMDraw
}

func(e *enemy) init() {
	e.step = 1
}

func(e *enemy) update(dt float64, targetPos pixel.Vec) {
	e.counter += dt

	// If we reached the target assign a new one (updated character position)
	if ((int(e.pos.X) >= int(e.target.X-2)) && (int(e.pos.X) <= int(e.target.X+2))) &&
	((int(e.pos.Y) >= int(e.target.Y-2)) && (int(e.pos.Y) <= int(e.target.Y+2))) {
		rand := rand2.Float64()	* 10

		e.initialPos = e.pos
		e.target = targetPos
		e.midPoint = pixel.Vec{X: e.initialPos.X + (targetPos.X-e.initialPos.X)/2, Y: (e.initialPos.Y + (targetPos.Y-e.initialPos.Y)/2)+rand}

		e.counter = 0
	}

	// Bezier curve lerp towards target
	m1 := pixel.Lerp(e.initialPos, e.midPoint, e.counter)
	m2 := pixel.Lerp(e.midPoint, e.target, e.counter)

	e.pos = pixel.Lerp(m1, m2, e.counter)
}

func(e *enemy) draw(t pixel.Target) {
	e.imd.Clear()

	e.imd.Color = e.color1

	e.imd.Push(pixel.V(e.pos.X, e.pos.Y))
	e.imd.Push(pixel.V(e.pos.X-5, e.pos.Y+0))
	e.imd.Push(pixel.V(e.pos.X-5, e.pos.Y+1))
	e.imd.Push(pixel.V(e.pos.X-2.5, e.pos.Y+1))

	e.imd.Color = e.color2

	e.imd.Push(pixel.V(e.pos.X-2.5, e.pos.Y+5))
	e.imd.Push(pixel.V(e.pos.X+2.5, e.pos.Y+5))
	e.imd.Push(pixel.V(e.pos.X+2.5, e.pos.Y+1))
	e.imd.Push(pixel.V(e.pos.X+5, e.pos.Y+1))
	e.imd.Push(pixel.V(e.pos.X+5, e.pos.Y+0))
	e.imd.Polygon(0)

	e.imd.Draw(t)
}