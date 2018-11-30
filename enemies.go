package main

import (
	"math"
	"path/filepath"
	"time"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
)

const (
	maxNumberOfEnemies = 20
)

var (
	attackBuildUpDuration = time.Millisecond * 400
)

type enemiesCollection struct {
	enemies []*enemy

	counter    float64
	step       float64
	difficulty float64

	img pixel.Picture
}

func (e *enemiesCollection) spawnPosition() pixel.Vec {

	container := streetBoundingRect.Norm()
	container.Min.X += 600
	container.Min.Y += 600
	container.Max.X -= 600
	container.Max.Y -= 600

	return randomPointInRect(container)
}

func (e *enemiesCollection) newEnemy() *enemy {
	return &enemy{
		spawnPos:  e.spawnPosition(),
		moveSpeed: 2,
		health:    100,
		maxHealth: 100,

		rect: pixel.R(-84, -74, 84, 74),

		color1: randomNiceColor(),
		color2: randomNiceColor(),

		imd: imdraw.New(nil),
		img: e.img,
	}
}

func (e *enemiesCollection) init() {
	var err error

	e.img, err = loadPicture("images/sprites/reaper")
	if err != nil {
		panic(err)
	}

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
	if len(e.enemies) <= maxNumberOfEnemies {
		if e.counter > e.step && characterIsOutside {
			enemy := e.newEnemy()

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

func (e *enemiesCollection) destroy() {

	for _, enemy := range e.enemies {
		enemy.die()
	}

	e.enemies = make([]*enemy, 0)

}

type enemy struct {
	vel       pixel.Vec
	moveSpeed float64

	rect              pixel.Rect
	health, maxHealth float64
	ded               bool
	isAttacking       bool

	spawnPos pixel.Vec
	target   pixel.Vec

	counter float64
	step    float64

	color1 pixel.RGBA
	color2 pixel.RGBA

	img    pixel.Picture
	imd    *imdraw.IMDraw
	sprite *pixel.Sprite

	//anim
	sheet                  pixel.Picture
	frame                  pixel.Rect
	anims                  map[string][]pixel.Rect
	rate, animCounter, dir float64
	lastBuildupFrameIndex  int
	attackBuildUpTime      time.Time
	attackAngle            float64
	attackAngleModifier    float64

	scythe *pixel.Sprite
}

func (e *enemy) Vel() pixel.Vec {
	return e.vel
}

func (e *enemy) HandleCollision(x Collidable, collisionTime float64, normal pixel.Vec) {
	switch collidable := x.(type) {
	case *laser:
		e.health -= collidable.damage

		playerScore.incrementScore(collidable.damage)

		if e.health <= 0 {
			e.die()
		}
	case *wall, *character:
		e.stopMotionCollision(collisionTime, normal)
	}
}

func (e *enemy) stopMotionCollision(collisionTime float64, normal pixel.Vec) {
	if normal.Y == 0 {
		// collision in X. move back by c.body.vel (with a negated x)
		e.rect = e.rect.Moved(e.vel.ScaledXY(pixel.V(-1, 0)))
		e.vel = e.vel.ScaledXY(pixel.V(0, 1))
	} else {
		// collision in Y. move back by c.body.vel (with a negated y)
		e.rect = e.rect.Moved(e.vel.ScaledXY(pixel.V(0, -1)))
		e.vel = e.vel.ScaledXY(pixel.V(1, 0))
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
	e.rect = e.rect.Moved(e.spawnPos)

	e.sprite = pixel.NewSprite(nil, pixel.Rect{})
	e.attackAngle = -1.2

	if e.scythe == nil {
		im, err := loadPicture("images/scythe")

		if err != nil {
			panic(err)
		}

		e.scythe = pixel.NewSprite(im, im.Bounds())
	}
}

func (e *enemy) update(dt float64, targetPos pixel.Vec) {
	e.animCounter += dt
	e.target = targetPos

	e.counter += dt

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

	distanceToTarget := e.rect.Center().Sub(e.target).Len()

	if distanceToTarget < 200 {
		if e.lastBuildupFrameIndex != len(e.anims["AttackBuild"])-1 {
			// we're not at the end of the attack build up. load the next attack build up frame
			e.lastBuildupFrameIndex = int(math.Floor(e.counter/e.rate)) % len(e.anims["AttackBuild"])
			e.frame = e.anims["AttackBuild"][e.lastBuildupFrameIndex]

			e.attackBuildUpTime = time.Now()
		} else {
			if time.Now().Sub(e.attackBuildUpTime) > attackBuildUpDuration {
				// we reached the end of the build up
				e.frame = e.anims["Attack"][0]
				e.isAttacking = true

				if e.attackAngle > -4.5 {
					e.attackAngle -= 0.1
				} else {
					e.clearAttackingState()
				}
			}
		}
	} else {
		e.clearAttackingState()
	}
}

func (e *enemy) clearAttackingState() {
	e.attackAngle = 0
	e.isAttacking = false
	e.lastBuildupFrameIndex = 0
	// @TODO attacking animation
	i := int(math.Floor(e.counter / e.rate / 2))
	e.frame = e.anims["Norm"][i%len(e.anims["Norm"])]
}

func (e *enemy) draw(t pixel.Target) {
	h := e.health / e.maxHealth

	e.sprite.Set(e.sheet, e.frame)

	m := pixel.IM.Moved(e.rect.Center())

	if e.vel.X > 0 {
		m = m.ScaledXY(e.rect.Center(), pixel.V(-1, 1))
	}

	e.sprite.DrawColorMask(t, m, pixel.RGB(h, h, h))

	if e.isAttacking {
		e.scythe.DrawColorMask(t, m.Rotated(e.rect.Center().Sub(pixel.V(10, 0)), e.attackAngle+0.1).ScaledXY(e.rect.Center().Sub(pixel.V(10, 0)), pixel.V(1, -1)), pixel.RGBA{R: 0.5, G: 0.5, B: 0.5, A: 0.8})
		e.scythe.DrawColorMask(t, m.Rotated(e.rect.Center().Sub(pixel.V(10, 0)), e.attackAngle).ScaledXY(e.rect.Center().Sub(pixel.V(10, 0)), pixel.V(1, -1)), pixel.RGBA{R: 0.5, G: 0.5, B: 0.5, A: 0.8})
		e.scythe.DrawColorMask(t, m.Rotated(e.rect.Center().Sub(pixel.V(10, 0)), e.attackAngle-0.1).ScaledXY(e.rect.Center().Sub(pixel.V(10, 0)), pixel.V(1, -1)), pixel.RGBA{R: 0.5, G: 0.5, B: 0.5, A: 0.8})
	}
}
