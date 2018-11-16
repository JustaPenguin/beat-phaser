package main

import (
	"image/color"
	"math/rand"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
)

type world struct {
	character *character
	enemies   *enemiesCollection

	platforms []platform
	rain      *rain

	weather   *imdraw.IMDraw
	mainScene *imdraw.IMDraw
}

func (w *world) init() {
	w.character = &character{}
	w.enemies = &enemiesCollection{}
	w.character.init()
	w.mainScene = imdraw.New(nil)
	w.weather = imdraw.New(nil)

	w.platforms = []platform{
		{rect: pixel.R(-50, -34, 50, -32)},
		{rect: pixel.R(20, 0, 70, 2)},
		{rect: pixel.R(-100, 10, -50, 12)},
		{rect: pixel.R(120, -22, 140, -20)},
		{rect: pixel.R(120, -72, 140, -70)},
		{rect: pixel.R(120, -122, 140, -120)},
		{rect: pixel.R(-100, -152, 100, -150)},
		{rect: pixel.R(-150, -127, -140, -125)},
		{rect: pixel.R(-180, -97, -170, -95)},
		{rect: pixel.R(-150, -67, -140, -65)},
		{rect: pixel.R(-180, -37, -170, -35)},
		{rect: pixel.R(-150, -7, -140, -5)},
	}
	for i := range w.platforms {
		w.platforms[i].color = randomNiceColor()
	}

	var rainDrops []pixel.Vec

	for i := 0; i < 1000; i++ {
		rainDrops = append(rainDrops, pixel.V((rand.Float64()*(win.Bounds().Max.X))-win.Bounds().Max.X/2, (rand.Float64()*(win.Bounds().Max.Y))-win.Bounds().Max.Y/2))
	}

	w.rain = &rain{
		positions: rainDrops,
	}
}

func (w *world) update(dt float64) {
	w.rain.update(w.character.body.rect.Center().Y-win.Bounds().Max.Y/2, w.character.body.rect.Center().Y+win.Bounds().Max.Y/2)
	w.character.update(dt)
	w.enemies.update(dt, w.character.body.rect.Center())
}

func (w *world) draw(t pixel.Target) {
	w.mainScene.Clear()
	w.weather.Clear()

	w.character.draw(t)
	w.enemies.draw(t)
	w.rain.draw(w.weather)

	for _, p := range w.platforms {
		p.draw(w.mainScene)
	}

	w.weather.Draw(t)
	w.mainScene.Draw(t)
}

type rain struct {
	positions []pixel.Vec
}

func (r *rain) update(lowerLimit, top float64) {
	xRange := rand.Float64() - 0.5

	for i := range r.positions {
		r.positions[i].Y -= rand.Float64()
		r.positions[i].X -= xRange

		if r.positions[i].Y < lowerLimit {
			r.positions[i].Y = top
		}
	}
}

func (r *rain) draw(imd *imdraw.IMDraw) {
	imd.Color = color.White

	for _, position := range r.positions {
		imd.Push(pixel.V(position.X, position.Y))
		imd.Push(pixel.V(position.X+1, position.Y))
		imd.Push(pixel.V(position.X+1, position.Y-1))
		imd.Push(pixel.V(position.X, position.Y-1))
		imd.Polygon(0)
	}
}

type platform struct {
	rect  pixel.Rect
	color color.Color
}

func (p *platform) draw(imd *imdraw.IMDraw) {
	imd.Color = p.color
	imd.Push(p.rect.Min, p.rect.Max)
	imd.Rectangle(0)
}
