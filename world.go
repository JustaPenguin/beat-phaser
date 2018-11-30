package main

import (
	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
	"golang.org/x/image/colornames"
	"image"
	"image/color"
	"math"
	"math/rand"
	"os"
	"path/filepath"
)

type world struct {
	character   *character
	enemies     *enemiesCollection
	advert      *advert
	deadMessage *deadMessage

	rain  *rain
	rooms []*room
	lights []*colorLight

	weather   *imdraw.IMDraw
	mainScene *imdraw.IMDraw
}

var ded bool
var healthDisplay float64
var wallMidpointPositionVec = pixel.V(0, -50)
var streetBoundingRect = pixel.R(-2100, -200, 2100, -2100)

func (w *world) init() {
	w.character = &character{}
	w.enemies = &enemiesCollection{}
	w.character.init()
	w.enemies.init()
	w.mainScene = imdraw.New(nil)
	w.weather = imdraw.New(nil)

	w.advert = &advert{
		pos:      pixel.V(-440, 155),
		maxWidth: 45,
	}
	w.advert.init()

	w.deadMessage = &deadMessage{
		pos: w.character.body.rect.Center().Add(pixel.V(-52, 50)),
	}
	w.deadMessage.init()

	w.rooms = append(w.rooms, &room{
		path: "/world-layer-background-bottom",
		walls: []*wall{
			// outside bounds
			{rect: pixel.R(-700, -200, 700, -190)}, // bottom outermost wall
			{rect: pixel.R(-710, 700, -700, -200)}, // left outermost wall
			{rect: pixel.R(700, 340, 710, -540)},   // right outermost wall
			{rect: pixel.R(700, 700, 710, 580)},
			{rect: pixel.R(-700, 690, 1110, 700).Moved(wallMidpointPositionVec)}, // top outermost wall

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

			{rect: pixel.R(260, 635, 0, 610)}, // kitchen top
			{rect: pixel.R(0, 610, 50, 545)},  // kitchen side
		},
	})

	w.rooms = append(w.rooms, &room{path: "/world-layer-background-top", topLayer: true, walls: []*wall{
		{rect: pixel.R(-710, 250, -620, 150)}, // plant
	}})
	w.rooms = append(w.rooms, &room{path: "/world-layer-animation", animLayer: true, rate: 1.0 / 10})

	// Wall with stairs layers
	w.rooms = append(w.rooms, &room{path: "/wall-stairs-layer-background-bottom", offset: pixel.V(1400, 0), walls: []*wall{
		{rect: pixel.R(910, 340, 925, -540)},   // stairs left hand wall
		{rect: pixel.R(1110, 655, 1120, -540)}, // stair room right hand wall
		{rect: pixel.R(930, -190, 1120, -200)}, // stair room base
	}})
	w.rooms = append(w.rooms, &room{path: "/wall-stairs-layer-background-top", offset: pixel.V(1400, 0), topLayer: true})

	// Wall layers
	w.rooms = append(w.rooms, &room{path: "/wall-layer-background-bottom", offset: pixel.V(-1400, 0)})
	w.rooms = append(w.rooms, &room{path: "/wall-layer-background-top", offset: pixel.V(-1400, 0), topLayer: true})

	// Street layers
	w.rooms = append(w.rooms, &room{path: "/street-base", offset: pixel.V(0, -1400), walls: []*wall{
		{rect: pixel.R(-2100, -540, 710, -550)}, // street top left
		{rect: pixel.R(910, -540, 2100, -550)},  // street top right

		{rect: pixel.R(streetBoundingRect.Min.X, streetBoundingRect.Max.Y, streetBoundingRect.Max.X, streetBoundingRect.Max.Y-10)}, // bottom
		{rect: pixel.R(streetBoundingRect.Min.X, streetBoundingRect.Max.Y, streetBoundingRect.Min.X-10, streetBoundingRect.Min.Y)}, // left
		{rect: pixel.R(streetBoundingRect.Max.X, streetBoundingRect.Max.Y, streetBoundingRect.Max.X+10, streetBoundingRect.Min.Y)}, // right
	}})
	w.rooms = append(w.rooms, &room{path: "/street-base", offset: pixel.V(1400, -1400)})
	w.rooms = append(w.rooms, &room{path: "/street-base", offset: pixel.V(-1400, -1400)})

	for _, room := range w.rooms {
		room.init(room.path)
	}

	w.rain = &rain{
		boundingRect: streetBoundingRect,
	}

	w.rain.init()

	w.lights = []*colorLight {
		{
			color:  colornames.Yellow,
			point:  pixel.V(-667, 695),
			angle:  -math.Pi / 2,
			radius: 50,
			spread: math.Pi / math.E,
		},
		{
			color:  colornames.Yellow,
			point:  pixel.V(-552, 695),
			angle:  -math.Pi / 2,
			radius: 50,
			spread: math.Pi / math.E,
		},
	}

	for _, light := range w.lights {
		light.init()
	}
}

func randomPointInRect(r pixel.Rect) pixel.Vec {
	base := r.Min

	return base.Add(pixel.V(r.W()*rand.Float64(), r.H()*rand.Float64()))
}

func (w *world) update(dt float64) {
	w.rain.update()
	w.character.update(dt)
	w.enemies.update(dt, w.character.body.rect.Center())
	w.advert.update(dt)
	w.deadMessage.update(dt, w.character.body.rect.Center().Add(pixel.V(-52, 50)))

	for _, room := range w.rooms {
		if room.animLayer {
			room.update(dt)
		}
	}
}

func (w *world) draw(t pixel.Target) {
	w.mainScene.Clear()
	w.weather.Clear()

	for _, room := range w.rooms {
		if !room.topLayer && !room.animLayer {
			room.drawnRoom.Draw(t)
		} else if room.animLayer {
			room.animDraw(t)
		}
	}

	w.character.draw(t)
	w.enemies.draw(t)


	for _, light := range w.lights {
		light.draw(t)
	}

	for _, room := range w.rooms {
		if room.topLayer && !room.animLayer {
			room.drawnRoom.Draw(t)
		}
	}

	//w.ui.draw

	w.rain.draw(w.weather)

	w.weather.Draw(t)

	w.mainScene.Draw(t)
	w.advert.draw(t)

	if ded {
		healthDisplay -= 0.01
		if healthDisplay <= 0 {
			healthDisplay = 0
		}

		imd := imdraw.New(nil)

		imd.Color = colornames.Black
		imd.Intensity = healthDisplay

		imd.Push(pixel.V(-10000, -10000))
		imd.Push(pixel.V(-10000, 10000))
		imd.Push(pixel.V(10000, 10000))
		imd.Push(pixel.V(10000, -10000))

		imd.Polygon(0)

		imd.Draw(t)

		w.character.draw(t)

		w.deadMessage.draw(t)

	} else {
		healthDisplay = 1
	}
}

type rain struct {
	positions    []pixel.Vec
	boundingRect pixel.Rect

	color color.Color
}

func (r *rain) init() {
	for i := 0; i < 2000; i++ {
		r.positions = append(r.positions, randomPointInRect(r.boundingRect))
	}
}

func (r *rain) update() {
	xRange := rand.Float64() - 0.5

	if playerScore.timeWindow {
		r.color = colornames.Blueviolet
	} else {
		r.color = colornames.Blue
	}

	for i := range r.positions {
		r.positions[i].Y -= rand.Float64()
		r.positions[i].X -= xRange

		if r.positions[i].Y < r.boundingRect.Max.Y {
			r.positions[i].Y = r.boundingRect.Min.Y
		}
	}
}

func (r *rain) draw(imd *imdraw.IMDraw) {
	imd.Color = r.color

	for _, position := range r.positions {
		imd.Push(pixel.V(position.X, position.Y))
		imd.Push(pixel.V(position.X, position.Y+5))
		imd.Polygon(1)
	}
}

type room struct {
	topLayer, animLayer bool
	path                string
	drawnRoom           *imdraw.IMDraw
	walls               []*wall

	img    pixel.Picture
	imd    *imdraw.IMDraw
	sprite *pixel.Sprite
	offset pixel.Vec

	//anim
	sheet         pixel.Picture
	anims         map[string][]pixel.Rect
	frame         pixel.Rect
	counter, rate float64
}

func (r *room) init(path string) {
	if r.animLayer {
		var err error

		r.sheet, r.anims, err = loadAnimationSheet("world-layer-animation", 1400, filepath.Join("images", "world", "rooms"))

		if err != nil {
			panic(err)
		}

		r.imd = imdraw.New(r.sheet)

		r.sprite = pixel.NewSprite(nil, pixel.Rect{})
	}

	for _, wall := range r.walls {
		wall.init()
	}

	if !r.animLayer {
		var err error

		r.img, err = loadPicture(filepath.Join("images", "world", "rooms") + path)
		if err != nil {
			panic(err)
		}

		r.sprite = pixel.NewSprite(r.img, r.img.Bounds())
		r.drawnRoom = imdraw.New(r.img)

		r.imd = imdraw.New(nil)

		r.draw(r.drawnRoom)
	}
}

func (r *room) update(dt float64) {
	r.counter += dt

	i := int(math.Floor(r.counter / r.rate / 2))
	r.frame = r.anims["Norm"][i%len(r.anims["Norm"])]
}

func (r *room) draw(t pixel.Target) {
	//r.image.Draw(t, pixel.IM.Scaled(r.image.Frame().Center(), 2.5))
	r.sprite.Draw(t, pixel.IM.Moved(r.offset))

	r.imd.Clear()

	for _, w := range r.walls {
		w.draw(r.imd)
	}

	r.imd.Draw(t)
}

func (r *room) animDraw(t pixel.Target) {
	r.imd.Clear()
	r.sprite.Set(r.sheet, r.frame)
	r.sprite.Draw(r.imd, pixel.IM)

	r.imd.Draw(t)
}

func loadPicture(path string) (pixel.Picture, error) {
	file, err := os.Open(path + ".png")
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
