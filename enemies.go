package main

import (
	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
	"math"
	"path/filepath"
)

type enemiesCollection struct {
	enemies []*enemy

	counter    float64
	step       float64
	difficulty float64

	img pixel.Picture
}

func (e *enemiesCollection) init() {
	var err error

	e.img, err = loadPicture("images/sprites/reaper")
	if err != nil {
		panic(err)
	}

	e.enemies = append(e.enemies, &enemy{
		initialPos: pixel.Vec{X: 0, Y: 0},
		moveSpeed:  2,
		health:     100,
		maxHealth:  100,

		rect: pixel.R(-84, -74, 84, 74),

		color1: randomNiceColor(),
		color2: randomNiceColor(),

		imd: imdraw.New(nil),
		img: e.img,
	})

	e.step = 2
	e.difficulty = 100

	for i := range e.enemies {
		e.enemies[i].init()
	}
}

func (e *enemiesCollection) update(dt float64, targetPos pixel.Vec) {
	e.counter += dt

	for i := len(e.enemies) - 1; i >= 0; i-- {
		// If enemy is ded remove from slice
		if e.enemies[i].ded {
			// For every ded enemy difficulty increases
			e.difficulty += 20
			e.enemies = e.enemies[:i+copy(e.enemies[i:], e.enemies[i+1:])]
		}
	}

	// e.step seconds have passed, add a new enemy (and increase spawn rate)
	// max enemies 50 (could be more but hey)
	if len(e.enemies) <= 50 {
		if e.counter > e.step {
			enemy := &enemy{
				initialPos: pixel.Vec{X: 0, Y: 0},
				moveSpeed:  2,
				health:     math.Round(e.difficulty/10) * 10, // health is nearest 20 to difficulty
				maxHealth:  math.Round(e.difficulty/10) * 10,

				rect: pixel.R(-84, -74, 84, 74),

				color1: randomNiceColor(),
				color2: randomNiceColor(),

				imd: imdraw.New(nil),
				img: e.img,
			}

			enemy.init()

			e.enemies = append(e.enemies, enemy)

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

func (e *enemiesCollection) draw(t pixel.Target) {
	for i := range e.enemies {
		e.enemies[i].draw(t)
	}
}

type enemy struct {
	vel       pixel.Vec
	moveSpeed float64

	rect              pixel.Rect
	health, maxHealth float64
	ded               bool

	initialPos pixel.Vec // Used for lerp, left just in case
	midPoint   pixel.Vec
	target     pixel.Vec

	counter float64
	step    float64

	color1 pixel.RGBA
	color2 pixel.RGBA

	img    pixel.Picture
	imd    *imdraw.IMDraw
	sprite *pixel.Sprite

	//anim
	sheet pixel.Picture
	frame pixel.Rect
	anims map[string][]pixel.Rect
	rate, animCounter, dir float64
}

func (e *enemy) Vel() pixel.Vec {
	return e.vel
}

func (e *enemy) HandleCollision(x Collidable, collisionTime float64, normal pixel.Vec) {
	switch collidable := x.(type) {
	case *laser:

		e.health -= collidable.damage

		if e.health <= 0 {
			e.die()
		}
	case *wall, *character, *enemy:
		e.stopMotionCollision(collisionTime, normal)
	}
}

func (e *enemy) stopMotionCollision(collisionTime float64, normal pixel.Vec) {
	if normal.Y == 0 {
		// collision in X. move back by c.body.vel (with a negated x)
		e.rect = e.rect.Moved(e.vel.ScaledXY(pixel.V(-1, 0)))
	} else {
		// collision in Y. move back by c.body.vel (with a negated y)
		e.rect = e.rect.Moved(e.vel.ScaledXY(pixel.V(0, -1)))
	}
}

func (e *enemy) Rect() pixel.Rect {
	return e.rect
}

func (e *enemy) die() {
	// @TODO play death animation?
	e.ded = true

	defer deregisterCollidable(e)
}

func (e *enemy) init() {
	defer registerCollidable(e)

	e.step = 1
	e.rate = 1.0 / 10

	var err error

	e.sheet, e.anims, err = loadAnimationSheet("reaper", 168, filepath.Join("images", "sprites"))

	if err != nil {
		panic(err)
	}

	e.imd = imdraw.New(e.sheet)

	e.sprite = pixel.NewSprite(nil, pixel.Rect{})
}

func (e *enemy) update(dt float64, targetPos pixel.Vec) {
	e.animCounter += dt

	// @TODO attacking animation
	i := int(math.Floor(e.counter / e.rate / 2))
	e.frame = e.anims["Norm"][i%len(e.anims["Norm"])]

	e.counter += dt

	// If we reached the target assign a new one (updated character position) OLD LERPING CODE
	/*if ((int(e.pos.X) >= int(e.target.X-2)) && (int(e.pos.X) <= int(e.target.X+2))) &&
		((int(e.pos.Y) >= int(e.target.Y-2)) && (int(e.pos.Y) <= int(e.target.Y+2))) {
		rand := rand2.Float64() * 10

		e.initialPos = e.pos
		e.target = targetPos
		e.midPoint = pixel.Vec{X: e.initialPos.X + (targetPos.X-e.initialPos.X)/2, Y: (e.initialPos.Y + (targetPos.Y-e.initialPos.Y)/2) + rand}

		e.counter = 0
	}

	// Bezier curve lerp towards target
	// @TODO control lerp speed?
	m1 := pixel.Lerp(e.initialPos, e.midPoint, e.counter)
	m2 := pixel.Lerp(e.midPoint, e.target, e.counter)

	e.pos = pixel.Lerp(m1, m2, e.counter)*/

	if e.rect.Center().X < targetPos.X {
		e.vel.X += dt
		if e.vel.X >= e.moveSpeed {
			e.vel.X = e.moveSpeed
		}
	}

	if e.rect.Center().X > targetPos.X {
		e.vel.X -= dt
		if e.vel.X <= -e.moveSpeed {
			e.vel.X = -e.moveSpeed
		}
	}

	if e.rect.Center().Y < targetPos.Y {
		e.vel.Y += dt
		if e.vel.Y >= e.moveSpeed {
			e.vel.Y = e.moveSpeed
		}
	}

	if e.rect.Center().Y > targetPos.Y {
		e.vel.Y -= dt
		if e.vel.Y <= -e.moveSpeed {
			e.vel.Y = -e.moveSpeed
		}
	}

	e.rect = e.rect.Moved(e.vel)
}

func (e *enemy) draw(t pixel.Target) {
	h := e.health / e.maxHealth

	e.imd.Clear()
	e.sprite.Set(e.sheet, e.frame)

	e.sprite.DrawColorMask(t, pixel.IM.Moved(e.rect.Center()), pixel.RGB(h, h, h))
}
