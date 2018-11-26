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
	speed: 500,
}

type weapon struct {
	lasers []*laser
	speed  float64

	matrix    pixel.Matrix
	parentPos pixel.Vec

	lastClick time.Time
	imdraw    *imdraw.IMDraw

	increment, multiplier int

	hitSplashes []*hitSplash

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
		thickness: 10,
	}
	l.init()

	w.lasers = append(w.lasers, l)
}

func (w *weapon) draw(t pixel.Target) {
	w.imdraw.Clear()
	w.imdraw.SetMatrix(pixel.IM)

	for _, laser := range w.lasers {
		laser.draw(w.imdraw)
	}

	for _, hitSplash := range w.hitSplashes {
		hitSplash.draw(w.imdraw)
	}

	w.imdraw.Draw(t)
}

func (w *weapon) update(dt float64, characterPos pixel.Vec, parentVelocity pixel.Vec) {
	timeSinceClick := time.Since(w.lastClick).Seconds()

	w.parentPos = characterPos.Add(pixel.V(5, 0))

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

		w.fire(characterPos, a, c)

		gameScore.setMultiplier(w.multiplier)

		go w.pringlePhaser.play()
	}

	for i := len(w.lasers) - 1; i >= 0; i-- {
		w.lasers[i].update(dt)

		if w.lasers[i].splash {
			w.hitSplashes = append(w.hitSplashes, &hitSplash{
				pos:    w.lasers[i].Rect().Center(),
				normal: w.lasers[i].splashNormal,
			})
		}

		if w.lasers[i].numCollisions > 3 || w.lasers[i].thickness <= 0 {
			w.lasers[i].destroy()
			w.lasers = w.lasers[:i+copy(w.lasers[i:], w.lasers[i+1:])]
		}
	}

	for i := len(w.hitSplashes) - 1; i >= 0; i-- {
		w.hitSplashes[i].update()

		// If enemy is ded remove from slice
		if w.hitSplashes[i].done {
			w.hitSplashes = w.hitSplashes[:i+copy(w.hitSplashes[i:], w.hitSplashes[i+1:])]
		}
	}
}

type laser struct {
	color        color.Color
	velocity     pixel.Vec
	lastVelocity pixel.Vec

	pos pixel.Vec

	thickness     float64
	numCollisions int
	splash        bool
	splashNormal  pixel.Vec
}

func (l *laser) init() {
	registerCollidable(l)
}

func (l *laser) destroy() {
	// increasing collisions arbitrarily causes removal of laser
	l.numCollisions = 4

	deregisterCollidable(l)
}

func (l *laser) HandleCollision(x Collidable, collisionTime float64, normal pixel.Vec) {
	switch x.(type) {
	case *laser, *character:
		return
	case *enemy:
		//@TODO create "hit" splash
		l.splash = true
		l.splashNormal = normal
		l.destroy()
	}

	if normal.X != 0 {
		l.velocity.X = -l.velocity.X
	}

	if normal.Y != 0 {
		l.velocity.Y = -l.velocity.Y
	}

	l.color = colornames.Hotpink
	l.numCollisions++
}

func (l *laser) Vel() pixel.Vec {
	return l.lastVelocity
}

func (l *laser) Rect() pixel.Rect {
	return pixel.R(l.pos.X-l.thickness, l.pos.Y-l.thickness, l.pos.X, l.pos.Y).Moved(pixel.V(l.thickness/2, l.thickness/2))
}

func (l *laser) update(dt float64) {
	l.lastVelocity = l.velocity.Scaled(dt)

	// move the position or expire the laser
	l.pos = l.pos.Add(l.lastVelocity)

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

type hitSplash struct {
	pos    pixel.Vec
	normal pixel.Vec

	x    float64
	done bool

	imd *imdraw.IMDraw
}

func (h *hitSplash) init() {

}

func (h *hitSplash) update() {
	h.x++

	if h.x > 10 {
		h.done = true
	}
}

func (h *hitSplash) draw(imd *imdraw.IMDraw) {
	if imd == nil {
		imd = imdraw.New(nil)
	}

	imd.Color = pixel.RGB(1, 1, 1)

	imd.Push(pixel.V(h.pos.X+h.x, h.pos.Y))
	imd.Push(pixel.V(h.pos.X, h.pos.Y+h.x))
	imd.Push(pixel.V(h.pos.X-h.x, h.pos.Y))
	imd.Push(pixel.V(h.pos.X, h.pos.Y-h.x))

	imd.Polygon(0)
}
