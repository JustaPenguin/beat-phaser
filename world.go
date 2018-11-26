package main

import (
	"golang.org/x/image/colornames"
	"image"
	"image/color"
	"math/rand"
	"os"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
)

type world struct {
	character *character
	enemies   *enemiesCollection

	rain  *rain
	rooms []*room

	weather   *imdraw.IMDraw
	mainScene *imdraw.IMDraw
}

var wallMidpointPositionVec = pixel.V(0, -50)

func (w *world) init() {
	w.character = &character{}
	w.enemies = &enemiesCollection{}
	w.character.init()
	w.enemies.init()
	w.mainScene = imdraw.New(nil)
	w.weather = imdraw.New(nil)

	w.rooms = append(w.rooms,
		&room{
			path: "images/world/rooms/world-layer-background-bottom.png",
			walls: []*wall{
				// outside bounds
				{rect: pixel.R(-700, -200, 700, -190)},                              // bottom outermost wall
				{rect: pixel.R(-710, 700, -700, -200)},                              // left outermost wall
				{rect: pixel.R(700, 700, 710, -200)},                                // right outermost wall
				{rect: pixel.R(-700, 690, 700, 700).Moved(wallMidpointPositionVec)}, // top outermost wall

				// room divisors - top rooms
				{rect: pixel.R(-10, 685, -5, 540).Moved(wallMidpointPositionVec)},    // hat room right wall
				{rect: pixel.R(-230, 690, -165, 540).Moved(wallMidpointPositionVec)}, // outside/inside horizontal boundary (top)
				{rect: pixel.R(-230, 310, -165, 200).Moved(wallMidpointPositionVec)}, // outside/inside horizontal boundary (bottom)
				{rect: pixel.R(-10, 310, -5, 200).Moved(wallMidpointPositionVec)},    // hat room right wall (bottom of gap)

				// room divisors - bottom rooms
				{rect: pixel.R(-700, 190, 0, 200).Moved(wallMidpointPositionVec)},   // horizontal room boundary to first door
				{rect: pixel.R(140, 190, 315, 200).Moved(wallMidpointPositionVec)},  // horizontal room boundary between doors
				{rect: pixel.R(455, 190, 700, 200).Moved(wallMidpointPositionVec)},  // horizontal room boundary to outer wall
				{rect: pixel.R(150, 190, 160, -140).Moved(wallMidpointPositionVec)}, // vertical boundary between bottom two rooms
			},
		},
		&room{
			path: "images/world/rooms/world-layer-anim.png",
		},
		&room{
			path:     "images/world/rooms/world-layer-background-top.png",
			topLayer: true,
			walls: []*wall{
				{rect: pixel.R(-710, 250, -620, 0)},
			},
		},
	)

	for _, room := range w.rooms {
		room.init(room.path)
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

	for _, room := range w.rooms {
		if !room.topLayer {
			room.drawnRoom.Draw(t)
		}
	}

	w.character.draw(t)
	w.enemies.draw(t)

	for _, room := range w.rooms {
		if room.topLayer {
			room.drawnRoom.Draw(t)
		}
	}

	//w.ui.draw

	w.rain.draw(w.weather)

	w.weather.Draw(t)
	w.mainScene.Draw(t)
}

type rain struct {
	positions []pixel.Vec

	color color.Color
}

func (r *rain) update(lowerLimit, top float64) {
	xRange := rand.Float64() - 0.5

	if playerScore.timeWindow {
		r.color = colornames.Blueviolet
	} else {
		r.color = colornames.Blue
	}

	for i := range r.positions {
		r.positions[i].Y -= rand.Float64()
		r.positions[i].X -= xRange

		if r.positions[i].Y < lowerLimit {
			r.positions[i].Y = top
		}
	}
}

func (r *rain) draw(imd *imdraw.IMDraw) {
	imd.Color = r.color

	for _, position := range r.positions {
		imd.Push(pixel.V(position.X, position.Y))
		imd.Push(pixel.V(position.X+1, position.Y))
		imd.Push(pixel.V(position.X+1, position.Y-1))
		imd.Push(pixel.V(position.X, position.Y-1))
		imd.Polygon(0)
	}
}

type room struct {
	topLayer  bool
	path      string
	drawnRoom *imdraw.IMDraw
	walls     []*wall

	img    pixel.Picture
	imd    *imdraw.IMDraw
	sprite *pixel.Sprite
}

func (r *room) init(path string) {
	var err error

	r.img, err = loadPicture(path)
	if err != nil {
		panic(err)
	}

	r.imd = imdraw.New(nil)

	for _, wall := range r.walls {
		wall.init()
	}

	r.sprite = pixel.NewSprite(r.img, r.img.Bounds())

	r.drawnRoom = imdraw.New(r.img)
	r.draw(r.drawnRoom)
}

func (r *room) draw(t pixel.Target) {
	//r.image.Draw(t, pixel.IM.Scaled(r.image.Frame().Center(), 2.5))
	r.sprite.Draw(t, pixel.IM)

	r.imd.Clear()

	for _, w := range r.walls {
		w.draw(r.imd)
	}

	r.imd.Draw(t)
}

func loadPicture(path string) (pixel.Picture, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	img, _, err := image.Decode(file)
	if err != nil {
		return nil, err
	}
	return pixel.PictureDataFromImage(img), nil
}

type wall struct {
	rect pixel.Rect
}

func (w *wall) init() {
	registerCollidable(w)
}

func (w *wall) update(dt float64) {

}

func (w *wall) draw(imd *imdraw.IMDraw) {

}

func (w *wall) Rect() pixel.Rect {
	return w.rect
}

func (w *wall) Vel() pixel.Vec {
	return pixel.ZV // walls don't move, dummy
}

func (w *wall) HandleCollision(c Collidable, collisionTime float64, normal pixel.Vec) {

}
