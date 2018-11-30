package main

import (
	"math"
	"path/filepath"
	"time"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
	"github.com/faiface/pixel/pixelgl"
)

type gopherAnimationState int

const (
	idle gopherAnimationState = iota
	running
	jumping
	shooting
	dying
)

type body struct {
	imd *imdraw.IMDraw

	// phys
	gravity   float64
	runSpeed  float64
	jumpSpeed float64

	rect pixel.Rect
	vel  pixel.Vec

	// anim
	sheet    pixel.Picture
	anims    map[string][]pixel.Rect
	rate     float64
	state    gopherAnimationState
	counter  float64
	dir      float64
	frame    pixel.Rect
	runFrame int
	sprite   *pixel.Sprite

	armMatrix        pixel.Matrix
	armSprite        *pixel.Sprite
	shootPos         pixel.Vec
	shootInitialised int

	rotationPoint pixel.Vec

	health, maxHealth, h float64

	ctrl pixel.Vec

	isDodging bool
	dodgeEnd  <-chan time.Time
	lastDodge time.Time
}

func (gp *body) init() {
	if gp.sheet == nil || gp.anims == nil {
		var err error

		gp.sheet, gp.anims, err = loadAnimationSheet("spike", 104, filepath.Join("images", "sprites"))

		if err != nil {
			panic(err)
		}
	}

	if gp.armSprite == nil {
		im, err := loadPicture("images/sprites/arm")

		if err != nil {
			panic(err)
		}

		gp.armSprite = pixel.NewSprite(im, im.Bounds())
	}

	gp.maxHealth = 100
	gp.health = gp.maxHealth
}

func (gp *body) update(dt float64) {

	if win.JustPressed(pixelgl.MouseButtonRight) && time.Now().Sub(gp.lastDodge) > time.Millisecond * 600 {
		gp.isDodging = true
		gp.dodgeEnd = time.Tick(time.Millisecond * 300)
	}

	// control the body with keys

	if !gp.isDodging {
		gp.ctrl = pixel.ZV

		if gp.health > 0 {
			if win.Pressed(pixelgl.KeyA) {
				gp.ctrl.X--
			}
			if win.Pressed(pixelgl.KeyD) {
				gp.ctrl.X++
			}
			if win.Pressed(pixelgl.KeyW) {
				gp.ctrl.Y++
			}
			if win.Pressed(pixelgl.KeyS) {
				gp.ctrl.Y--
			}
		}
	}

	select {
	case <-gp.dodgeEnd:
		gp.isDodging = false
		gp.lastDodge = time.Now()
		gp.dodgeEnd = nil
	default:
	}

	dodgeMultiplier := 1.0

	// apply controls
	switch {
	case gp.ctrl.X < 0:
		gp.vel.X = -gp.runSpeed
	case gp.ctrl.X > 0:
		gp.vel.X = +gp.runSpeed
	default:
		gp.vel.X = 0
	}

	switch {
	case gp.ctrl.Y < 0:
		gp.vel.Y = -gp.runSpeed
	case gp.ctrl.Y > 0:
		gp.vel.Y = +gp.runSpeed
	default:
		gp.vel.Y = 0
	}

	if gp.isDodging {
		dodgeMultiplier = 2
	}

	gp.vel = gp.vel.Scaled(dt).Scaled(dodgeMultiplier)

	// apply gravity and velocity
	gp.rect = gp.rect.Moved(gp.vel)

	gp.counter += dt

	// determine the new animation state
	var newState gopherAnimationState
	switch {
	case gp.vel.Len() == 0:
		newState = idle
	case gp.vel.Len() > 0:
		newState = running
	}

	if win.JustPressed(pixelgl.MouseButtonLeft) && gp.state != running {
		newState = shooting
		gp.shootInitialised = 10
	}

	if gp.shootInitialised > 0 {
		newState = shooting
		gp.shootInitialised--
	}

	if gp.health <= 0 {
		newState = dying
	}

	// reset the time counter if the state changed
	if gp.state != newState {
		gp.state = newState
		gp.counter = 0
	}

	if !gp.isDodging {

		// determine the correct animation frame
		switch gp.state {
		case idle:
			i := int(math.Floor(gp.counter / gp.rate / 2))
			gp.frame = gp.anims["Front"][i%len(gp.anims["Front"])]
		case running:
			gp.runFrame = int(math.Floor(gp.counter/gp.rate)) % len(gp.anims["Run"])
			gp.frame = gp.anims["Run"][gp.runFrame]
		case shooting:
			gp.frame = gp.anims["Run"][1]
		case dying:
			i := int(math.Floor(gp.counter / gp.rate))

			if i >= len(gp.anims["Die"])-1 {
				gp.frame = gp.anims["Die"][len(gp.anims["Die"])-1]
			} else {
				gp.frame = gp.anims["Die"][i]
			}
		case jumping:
			speed := gp.vel.Y
			i := int((-speed/gp.jumpSpeed + 1) / 2 * float64(len(gp.anims["Jump"])))
			if i < 0 {
				i = 0
			}
			if i >= len(gp.anims["Front"]) {
				i = len(gp.anims["Front"]) - 1
			}
			gp.frame = gp.anims["Front"][i]
		}
	}
	// set the facing direction of the body
	if gp.vel.X != 0 {
		if gp.vel.X > 0 {
			gp.dir = -1
		} else {
			gp.dir = +1
		}
	}

	gp.updateArm(dt)
}

// bobbing is a map of how much the arm/body move up by when running
var bobbing = map[int]float64{
	1:  0,
	2:  1,
	3:  2,
	4:  2,
	5:  1,
	6:  0,
	7:  0,
	8:  1,
	9:  2,
	10: 2,
	11: 1,
	12: 0,
}

func (gp *body) updateArm(dt float64) {
	pb := gp.armSprite.Picture().Bounds()

	armPos := gp.rect.Center().Add(pixel.V(40, 32+bobbing[gp.runFrame]))

	mouseAngleFromCenter := getMouseAngleFromCenter()

	// undo the angle correction performed by getMouseAngleFromCenter
	if win.MousePosition().X < win.Bounds().Center().X {
		mouseAngleFromCenter += math.Pi + 0.3
	} else {
		mouseAngleFromCenter -= 0.3
	}

	// if we're running in the opposite way to the direction we're shooting
	isShootingBackwards := gp.dir < 0 && win.MousePosition().X < win.Bounds().Center().X || gp.dir > 0 && win.MousePosition().X >= win.Bounds().Center().X

	if gp.dir < 0 {
		gp.rotationPoint = armPos.Add(pixel.V(pb.W()/2*gp.dir, 0))
	} else {
		gp.rotationPoint = armPos.Sub(pixel.V(pb.W()/2*gp.dir, 0))
	}

	// move the picture to the armPos and scale it by the direction
	gp.armMatrix = pixel.IM.Moved(armPos).ScaledXY(gp.rect.Center(), pixel.V(-gp.dir, 1))

	if isShootingBackwards {
		gp.armMatrix = gp.armMatrix.ScaledXY(armPos, pixel.V(-1, -1)).Moved(pixel.V(-80, 0))

		if (mouseAngleFromCenter > math.Pi && gp.dir < 0) || (mouseAngleFromCenter < math.Pi && gp.dir > 0) {
			// flip the arm in Y if we're in a certain angle so the arm looks correct.
			gp.armMatrix = gp.armMatrix.ScaledXY(armPos, pixel.V(1, -1))
		}
	}

	// rotate by the mouse angle from the center, around the calculated rotation point
	gp.armMatrix = gp.armMatrix.Chained(pixel.IM.Rotated(gp.rotationPoint, mouseAngleFromCenter))

	// shootpos is the end of the gun
	gp.shootPos = gp.armMatrix.Project(pixel.ZV.Add(pixel.V(pb.W()/2, 2)))
}

func (gp *body) draw(t pixel.Target) {
	if gp.imd == nil {
		gp.imd = imdraw.New(gp.sheet)
	}

	gp.imd.Clear()

	if ded {
		gp.h += 0.05

		if gp.h >= 0.5 {
			gp.h = 0.8
		}
	} else {
		gp.h = gp.health / gp.maxHealth
	}

	if gp.state != idle && gp.state != dying {
		// only draw the arm if we're not idling
		gp.armSprite.DrawColorMask(t, gp.armMatrix, pixel.RGB(gp.h, gp.h, gp.h))
	}

	if gp.sprite == nil {
		gp.sprite = pixel.NewSprite(nil, pixel.Rect{})
	}

	// draw the correct frame with the correct position and direction
	gp.sprite.Set(gp.sheet, gp.frame)
	gp.sprite.DrawColorMask(gp.imd, pixel.IM.
		ScaledXY(pixel.ZV, pixel.V(
			gp.rect.W()/gp.sprite.Frame().W(),
			gp.rect.H()/gp.sprite.Frame().H(),
		)).
		ScaledXY(pixel.ZV, pixel.V(-gp.dir, 1)).
		Moved(gp.rect.Center()),
		pixel.RGB(gp.h, gp.h, gp.h),
	)
	gp.imd.Draw(t)
}

func todegrees(rads float64) float64 {
	return rads * (180 / math.Pi)
}

func angleBetweenVectors(v1, v2 pixel.Vec) float64 {
	angle := math.Atan2(v2.Y, v2.X) - math.Atan2(v1.Y, v2.X)

	if angle < 0 {
		angle += 2 * math.Pi
	}

	return angle
}

func getMouseAngleFromCenter() float64 {
	a := angleBetweenVectors(pixel.V(0, 0), win.Bounds().Center().Sub(win.MousePosition()))

	if win.MousePosition().X < win.Bounds().Center().X {
		a -= math.Pi
	}

	return a
}
