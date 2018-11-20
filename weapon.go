package main

import (
	"image/color"
	"time"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
	"github.com/faiface/pixel/pixelgl"
	"golang.org/x/image/colornames"
)

// handgun is a simple handgun-like weapon
var handgun = &weapon{
	speed: 100,
	points: []pixel.Vec{
		pixel.V(0, 0),
		pixel.V(1, 0),
		pixel.V(2, 0),
		pixel.V(2, 2),
		pixel.V(3, 2),
		pixel.V(4, 2),
		pixel.V(5, 2),
	},
}

type weapon struct {
	lasers []*laser
	speed  float64
	points []pixel.Vec

	matrix    pixel.Matrix
	parentPos pixel.Vec

	lastClick time.Time
	imdraw    *imdraw.IMDraw

	increment, multiplier int

	// sounds
	pringlePhaser *soundEffect
}

func (w *weapon) init() {
	w.lastClick = time.Now()

	if w.imdraw == nil {
		w.imdraw = imdraw.New(nil)
	}

	if w.pringlePhaser == nil {
		w.pringlePhaser = &soundEffect{
			filePath: "audio/effects/pringle-phaser.ogg",
		}
		w.pringlePhaser.load()
	}
}

func (w *weapon) fire(origin pixel.Vec, angle float64, color color.Color) {
	l := &laser{
		color:     color,
		velocity:  pixel.V(w.speed, 0).Rotated(angle),
		pos:       origin,
		thickness: 2,
	}
	l.init()

	w.lasers = append(w.lasers, l)
}

func (w *weapon) draw(t pixel.Target) {
	w.imdraw.Clear()
	w.imdraw.SetMatrix(w.matrix)

	w.imdraw.Color = colornames.Blueviolet

	for _, pt := range w.points {
		w.imdraw.Push(pt.Add(w.parentPos))
	}

	w.imdraw.Polygon(1)
	w.imdraw.SetMatrix(pixel.IM)

	for _, laser := range w.lasers {
		laser.draw(w.imdraw)
	}

	w.imdraw.Draw(t)
}

func (w *weapon) update(dt float64, characterPos pixel.Vec, parentVelocity pixel.Vec) {
	timeSinceClick := time.Since(w.lastClick).Seconds()

	w.parentPos = characterPos.Add(pixel.V(5, 0))
	w.matrix = pixel.IM.Rotated(w.points[0].Add(characterPos), getMouseAngleFromCenter())

	if parentVelocity.X < 0 && win.MousePosition().X < win.Bounds().Center().X {
		w.matrix = w.matrix.Scaled(characterPos, -1).Moved(pixel.V(-10, 0))
	}

	if win.JustPressed(pixelgl.MouseButtonLeft) {
		w.lastClick = time.Now()

		a := getMouseAngleFromCenter()

		var c color.Color

		if timeSinceClick > 60/bpm+0.05 || timeSinceClick < 60/bpm-0.05 {
			w.increment--

			if w.increment <= 0 {
				w.multiplier--
				w.increment = 8
			}

			if w.multiplier < 0 {
				w.multiplier = 0
			}

			c = color.White
		} else {
			w.increment++

			if w.increment >= 8 {
				w.multiplier++
				w.increment = 0
			}

			if w.multiplier > 8 {
				w.multiplier = 8
			}

			c = randomNiceColor()
		}

		w.fire(w.matrix.Project(characterPos.Add(w.points[len(w.points)-1])), a, c)

		gameScore.setMultiplier(w.multiplier)

		go w.pringlePhaser.play()
	}

	var toRemove []int

	for i, laser := range w.lasers {
		laser.update(dt)

		if laser.numCollisions > 3 || laser.thickness <= 0 {
			toRemove = append(toRemove, i)
		}
	}

	for _, i := range toRemove {
		w.lasers[i].destroy()
		w.lasers = append(w.lasers[:i], w.lasers[i+1:]...)
	}
}

type laser struct {
	color    color.Color
	velocity pixel.Vec

	pos, prevPos pixel.Vec

	thickness     float64
	numCollisions int
}

func (l *laser) init() {
	RegisterCollidable(l)
}

func (l *laser) destroy() {
	DeregisterCollidable(l)
}

func (l *laser) HandleCollision(x Collidable) {
	switch x.(type) {
	case *laser, *character:
		return
	}

	r := x.Rect()

	if int(l.prevPos.Y) <= int(r.Min.Y) || int(l.prevPos.Y) >= int(r.Max.Y) {
		l.velocity.Y = -l.velocity.Y
	} else {
		l.velocity.X = -l.velocity.X
	}

	l.color = colornames.Hotpink
	l.numCollisions++
}

func (l *laser) Rect() pixel.Rect {
	return pixel.R(l.pos.X - 1, l.pos.Y -1, l.pos.X, l.pos.Y)
}

func (l *laser) update(dt float64) {
	l.prevPos = l.pos
	// move the position or expire the laser
	l.pos = l.pos.Add(l.velocity.Scaled(dt))

	if l.thickness > 0 {
		l.thickness = l.thickness - 0.02
	}
}

func (l *laser) draw(imd *imdraw.IMDraw) {
	imd.Color = l.color
	imd.EndShape = imdraw.RoundEndShape

	imd.Push(l.pos)
	imd.Polygon(l.thickness)
}
