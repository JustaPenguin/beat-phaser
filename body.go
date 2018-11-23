package main

import (
	"fmt"
	"golang.org/x/image/colornames"
	"math"

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

	leftNormal string = "left"
	rightNormal string = "right"
	upNormal string = "up"
	downNormal string = "down"
)

var normalMap map[string]pixel.Vec

type body struct {
	imd, xImd *imdraw.IMDraw

	// phys
	gravity   float64
	runSpeed  float64
	jumpSpeed float64

	rect pixel.Rect
	vel  pixel.Vec

	colliding map[pixel.Vec]Collidable

	// anim
	sheet   pixel.Picture
	anims   map[string][]pixel.Rect
	rate    float64
	state   gopherAnimationState
	counter float64
	dir     float64
	frame   pixel.Rect
	sprite  *pixel.Sprite

	armMatrix pixel.Matrix
	armSprite *pixel.Sprite
	shootPos pixel.Vec
	shootInitialised int

	rotationPoint pixel.Vec
}

func (gp *body) init() {
	if normalMap == nil {
		normalMap = make(map[string]pixel.Vec)

		normalMap[leftNormal] = pixel.V(1, 0)
		normalMap[rightNormal] = pixel.V(-1, 0)
		normalMap[upNormal] = pixel.V(0, -1)
		normalMap[downNormal] = pixel.V(0, 1)
	}

	if gp.sheet == nil || gp.anims == nil {
		var err error

		gp.sheet, gp.anims, err = loadAnimationSheet("spike", 104)

		if err != nil {
			panic(err)
		}
	}

	if gp.colliding == nil {
		gp.colliding = make(map[pixel.Vec]Collidable)
	}

	if gp.armSprite == nil {
		im, err := loadPicture("images/sprites/arm.png")

		if err != nil {
			panic(err)
		}

		gp.armSprite = pixel.NewSprite(im, im.Bounds())
	}

	if gp.xImd == nil {
		gp.xImd = imdraw.New(nil)
	}
}

func (gp *body) checkCollisions(x Collidable, normal pixel.Vec) {
	switch x.(type) {
	case *laser, *character:
		return
	case *wall:
		// Moving the wall box by the normal scaled gives a bit of extra leeway before the collision is considered over
		// This makes spamming through walls more difficult, but not impossible
		if gp.rect.Norm().Intersect(x.Rect().Moved(normal.Scaled(20)).Norm()) != pixel.R(0,0,0,0) {
			// still collided
		} else {
			// no longer collided
			gp.colliding[normal] = nil
		}
	}
}

func (gp *body) update(dt float64) {

	for _, normal := range normalMap {
		if gp.colliding[normal] != nil {
			gp.checkCollisions(gp.colliding[normal], normal)
		}
	}

	// control the body with keys
	ctrl := pixel.ZV

	// This is a bit messy, but stops ctrl from being altered in a certain direction if the body
	// has collided in said direction
	var skipx bool
	var skipy bool

	if win.Pressed(pixelgl.KeyA) && gp.colliding[normalMap[leftNormal]] == nil {
		ctrl.X--
		skipx = true
	} else if gp.colliding[normalMap[leftNormal]] != nil {
		ctrl.X = 0
	}
	if win.Pressed(pixelgl.KeyD) && gp.colliding[normalMap[rightNormal]] == nil {
		ctrl.X++
	} else if gp.colliding[normalMap[rightNormal]] != nil && !skipx {
		ctrl.X = 0
	}
	if win.Pressed(pixelgl.KeyW) && gp.colliding[normalMap[upNormal]] == nil {
		ctrl.Y++
		skipy = true
	} else if gp.colliding[normalMap[upNormal]] != nil {
		ctrl.Y = 0
	}
	if win.Pressed(pixelgl.KeyS) && gp.colliding[normalMap[downNormal]] == nil {
		ctrl.Y--
	} else if gp.colliding[normalMap[downNormal]] != nil && !skipy {
		ctrl.Y = 0
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

	gp.vel = gp.vel.Scaled(dt)

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

	newState = running // @TODO REMOVE

	if win.JustPressed(pixelgl.MouseButtonLeft) && gp.state != running {
		newState = shooting
		gp.shootInitialised = 10
	}

	if gp.shootInitialised > 0 {
		newState = shooting
		gp.shootInitialised--
	}


	// reset the time counter if the state changed
	if gp.state != newState {
		gp.state = newState
		gp.counter = 0
	}

	// determine the correct animation frame
	switch gp.state {
	case idle:
		i := int(math.Floor(gp.counter / gp.rate / 2))
		gp.frame = gp.anims["Front"][i%len(gp.anims["Front"])]
	case running:
		i := int(math.Floor(gp.counter / gp.rate))
		gp.frame = gp.anims["Run"][i%len(gp.anims["Run"])]
	case shooting:
		gp.frame = gp.anims["Run"][1]
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

func (gp *body) updateArm(dt float64) {
	pb := gp.armSprite.Picture().Bounds()
	pb.Max = pb.Max.Rotated(getMouseAngleFromCenter())

	// draw arm
	armPos := gp.rect.Center().Add(pixel.V(40, 40))

	a := getMouseAngleFromCenter()

	fmt.Println(a)
/*
	if a > 0.71 {
		a = 0.71
	}
*/
	if win.MousePosition().X < win.Bounds().Center().X {
		a += math.Pi
	}


	gp.armMatrix = pixel.IM.Moved(armPos). // move the picture to the armPos
		ScaledXY(gp.rect.Center(), pixel.V(-gp.dir, 1)). // scale it by the direction
		Rotated(armPos.Add(pixel.V(pb.W()/2*gp.dir + 1000, 1000)),a)



	gp.rotationPoint = gp.armMatrix.Project(pixel.ZV)

	// if we're running in the opposite way to the direction we're shooting, invert gp.dir here
	if gp.dir < 0 && win.MousePosition().X < win.Bounds().Center().X || gp.dir > 0 && win.MousePosition().X >= win.Bounds().Center().X  {
		gp.armMatrix = gp.armMatrix.ScaledXY(armPos, pixel.V(-1, 1)).Moved(pixel.V(-80, 0))
	}

	gp.shootPos = gp.armMatrix.Project(pixel.ZV).Add(pixel.V(pb.W()/2, 0))
}

func (gp *body) draw(t pixel.Target) {
	if gp.imd == nil {
		gp.imd = imdraw.New(gp.sheet)
	}

	gp.imd.Clear()

	if gp.state != idle {
		// only draw the arm if we're not idling
		gp.armSprite.Draw(t, gp.armMatrix)
	}

	if gp.sprite == nil {
		gp.sprite = pixel.NewSprite(nil, pixel.Rect{})
	}
	// draw the correct frame with the correct position and direction
	gp.sprite.Set(gp.sheet, gp.frame)
	gp.sprite.Draw(gp.imd, pixel.IM.
		ScaledXY(pixel.ZV, pixel.V(
			gp.rect.W()/gp.sprite.Frame().W(),
			gp.rect.H()/gp.sprite.Frame().H(),
		)).
		ScaledXY(pixel.ZV, pixel.V(-gp.dir, 1)).
		Moved(gp.rect.Center()),
	)
	gp.imd.Draw(t)

	gp.xImd.Clear()

	gp.xImd.Color = colornames.Cyan

	gp.xImd.Push(gp.rotationPoint)
	gp.xImd.Push(pixel.V(0,0))
	gp.xImd.Polygon(10)
	gp.xImd.Draw(t)
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
