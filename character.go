package main

import (
	"fmt"
	"golang.org/x/image/colornames"
	"image/color"
	"math"
	"time"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
	"github.com/faiface/pixel/pixelgl"
)

type character struct {
	body *body
	hat  *hat
	weapon *weapon
}

func (c *character) init() {

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

	c.weapon = &weapon{
		speed: 3,
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
	c.weapon.init()
}

func (c *character) update(dt float64) {
	c.body.update(dt)
	c.hat.update(dt, c.body.rect.Center())
	c.weapon.update(dt, c.body.rect.Center().Add(pixel.V(8, 2)))
}

func (c *character) draw(t pixel.Target) {
	c.body.draw(t)
	c.hat.draw(t)
	c.weapon.draw(t, c.body.rect.Center().Add(pixel.V(5, 0)))
}

type hat struct {
	pos             pixel.Vec
	counter         float64
	color, altColor pixel.RGBA
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

	// @TODO collisions with stuff looks like this, turn platforms into walls
	/*if gp.vel.Y <= 0 {
		for _, p := range platforms {
			if gp.rect.Max.X <= p.rect.Min.X || gp.rect.Min.X >= p.rect.Max.X {
				continue
			}
			if gp.rect.Min.Y > p.rect.Max.Y || gp.rect.Min.Y < p.rect.Max.Y+gp.vel.Y*dt {
				continue
			}
			gp.vel.Y = 0
			gp.rect = gp.rect.Moved(pixel.V(0, p.rect.Max.Y-gp.rect.Min.Y))
		}
	}*/

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

type laser struct {
	color color.Color
	expired bool
	angle float64

	velocity pixel.Vec

	pos pixel.Vec

	thickness float64
}

func (l *laser) update(dt float64) {
	// move the position or expire the laser
	l.pos = l.pos.Add(l.velocity.Rotated(l.angle))

	if l.thickness > 0 {
		l.thickness = l.thickness - 0.02
	}
}

func (l *laser) draw(imd *imdraw.IMDraw) {
	imd.Color = l.color
	imd.EndShape = imdraw.RoundEndShape

	imd.Push(l.pos, l.pos.Add(pixel.V(5, 0).Rotated(l.angle)))
	imd.Line(l.thickness)
}

type weapon struct {
	lasers []*laser
	speed float64
	points []pixel.Vec

	lastClick             time.Time
	imdraw *imdraw.IMDraw

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
	w.lasers = append(w.lasers, &laser{
		color: color,
		velocity: pixel.V(w.speed, 0),
		angle: angle,
		pos: origin,
		// Minus half window size
		thickness: 2,
	})
}

func (w *weapon) draw(t pixel.Target, parent pixel.Vec) {
	w.imdraw.Clear()

	w.imdraw.Color = colornames.Blueviolet

	for _, pt := range w.points {
		w.imdraw.Push(pt.Add(parent))
	}

	w.imdraw.Polygon(1)

	for _, laser := range w.lasers {
		laser.draw(w.imdraw)
	}

	w.imdraw.Draw(t)
}

func todegrees(rads float64) float64 {
	return rads * (180/math.Pi)
}

func angleBetweenVectors(v1, v2 pixel.Vec) float64 {
	angle := math.Atan2(v2.Y, v2.X) - math.Atan2(v1.Y, v2.X)

	if angle < 0 {
		angle += 2 * math.Pi
	}

	return angle
}

func (w *weapon) update(dt float64, characterPos pixel.Vec) {
	timeSinceClick := time.Since(w.lastClick).Seconds()

	if win.JustPressed(pixelgl.MouseButtonLeft) {
		w.lastClick = time.Now()

		println("a", camPos.Angle())
		//a := win.MousePosition().Angle() - characterPos.Add(win.Bounds().Center()).Angle()
		fmt.Println(win.MousePosition())
		fmt.Println(win.Bounds().Center())
		a := angleBetweenVectors(pixel.V(0, 0), win.Bounds().Center().Sub(win.MousePosition()))

		if win.MousePosition().X < win.Bounds().Center().X {
			a -= math.Pi
		}

		//a := angleBetweenVectors(win.Bounds().Center(), win.Bounds().Center().Sub(win.MousePosition()))

		fmt.Println(todegrees(a))

		if timeSinceClick > 60/bpm+0.05 || timeSinceClick < 60/bpm-0.05 {
			w.increment--

			if w.increment <= 0 {
				w.multiplier--
				w.increment = 8
			}

			if w.multiplier < 0 {
				w.multiplier = 0
			}

			w.fire(characterPos, a, color.White)
		} else {
			w.increment++

			if w.increment >= 8 {
				w.multiplier++
				w.increment = 0
			}

			if w.multiplier > 8 {
				w.multiplier = 8
			}

			w.fire(characterPos, a, randomNiceColor())
		}

		gameScore.setMultiplier(w.multiplier)

		go w.pringlePhaser.play()
	}

	var expiredLasers []int

	for i, laser := range w.lasers {
		laser.update(dt)

		if laser.expired {
			expiredLasers = append(expiredLasers, i)
		}
	}
}