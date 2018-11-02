package main

import (
	"encoding/csv"
	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
	"github.com/faiface/pixel/pixelgl"
	"github.com/pkg/errors"
	"golang.org/x/image/colornames"
	"image"
	"image/color"
	_ "image/png"
	"io"
	"math"
	"math/rand"
	"os"
	"strconv"
	"time"
)

func loadAnimationSheet(sheetPath, descPath string, frameWidth float64) (sheet pixel.Picture, anims map[string][]pixel.Rect, err error) {
	// total hack, nicely format the error at the end, so I don't have to type it every time
	defer func() {
		if err != nil {
			err = errors.Wrap(err, "error loading animation sheet")
		}
	}()

	// open and load the spritesheet
	sheetFile, err := os.Open(sheetPath)
	if err != nil {
		return nil, nil, err
	}
	defer sheetFile.Close()
	sheetImg, _, err := image.Decode(sheetFile)
	if err != nil {
		return nil, nil, err
	}
	sheet = pixel.PictureDataFromImage(sheetImg)

	// create a slice of frames inside the spritesheet
	var frames []pixel.Rect
	for x := 0.0; x+frameWidth <= sheet.Bounds().Max.X; x += frameWidth {
		frames = append(frames, pixel.R(
			x,
			0,
			x+frameWidth,
			sheet.Bounds().H(),
		))
	}

	descFile, err := os.Open(descPath)
	if err != nil {
		return nil, nil, err
	}
	defer descFile.Close()

	anims = make(map[string][]pixel.Rect)

	// load the animation information, name and interval inside the spritesheet
	desc := csv.NewReader(descFile)
	for {
		anim, err := desc.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, nil, err
		}

		name := anim[0]
		start, _ := strconv.Atoi(anim[1])
		end, _ := strconv.Atoi(anim[2])

		anims[name] = frames[start : end+1]
	}

	return sheet, anims, nil
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

type laser struct {
	rect pixel.Rect
	color color.Color

	thickness float64
}

func (l *laser) draw(imd *imdraw.IMDraw) {
	imd.Color = l.color
	imd.EndShape = imdraw.RoundEndShape

	imd.Push(pixel.V(l.rect.Min.X, l.rect.Min.Y), pixel.V(l.rect.Max.X, l.rect.Max.Y))
	imd.Line(l.thickness)

	if l.thickness > 0 {
		l.thickness = l.thickness - 0.02
	}
}

type gopherPhys struct {
	gravity   float64
	runSpeed  float64
	jumpSpeed float64

	rect   pixel.Rect
	vel    pixel.Vec
	ground bool
}

func (gp *gopherPhys) update(dt float64, ctrl pixel.Vec, platforms []platform) {
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
		gp.vel.Y= -gp.runSpeed
	case ctrl.Y > 0:
		gp.vel.Y = +gp.runSpeed
	default:
		gp.vel.Y = 0
	}

	// apply gravity and velocity
	gp.rect = gp.rect.Moved(gp.vel.Scaled(dt))

	// @TODO collisions with stuff looks like this
	/*gp.ground = false
	if gp.vel.Y <= 0 {
		for _, p := range platforms {
			if gp.rect.Max.X <= p.rect.Min.X || gp.rect.Min.X >= p.rect.Max.X {
				continue
			}
			if gp.rect.Min.Y > p.rect.Max.Y || gp.rect.Min.Y < p.rect.Max.Y+gp.vel.Y*dt {
				continue
			}
			gp.vel.Y = 0
			gp.rect = gp.rect.Moved(pixel.V(0, p.rect.Max.Y-gp.rect.Min.Y))
			gp.ground = true
		}
	}

	// jump if on the ground and the player wants to jump
	if gp.ground && ctrl.Y > 0 {
		gp.vel.Y = gp.jumpSpeed
	}*/
}

type fire struct {
	speed float64
	origin pixel.Vec
	vector pixel.Vec

	newLaser *laser
}

func (f *fire) now(speed float64, origin pixel.Vec, vector pixel.Vec) {
	f.newLaser = &laser{
		color: randomNiceColor(),
		// Minus half window size
		rect: pixel.R(origin.X-512, origin.Y-384, vector.X-512, vector.Y-384),
		thickness: 2,
	}
}

func (f *fire) draw(imd *imdraw.IMDraw) {
	if f.newLaser != nil {
		f.newLaser.draw(imd)
	}
}

type animState int

const (
	idle animState = iota
	running
	jumping
)

type gopherAnim struct {
	sheet pixel.Picture
	anims map[string][]pixel.Rect
	rate  float64

	state   animState
	counter float64
	dir     float64

	frame pixel.Rect

	sprite *pixel.Sprite
}

func (ga *gopherAnim) update(dt float64, phys *gopherPhys) {
	ga.counter += dt

	// determine the new animation state
	var newState animState
	switch {
	case !phys.ground:
		newState = jumping
	case phys.vel.Len() == 0:
		newState = idle
	case phys.vel.Len() > 0:
		newState = running
	}

	// reset the time counter if the state changed
	if ga.state != newState {
		ga.state = newState
		ga.counter = 0
	}

	// determine the correct animation frame
	switch ga.state {
	case idle:
		ga.frame = ga.anims["Front"][0]
	case running:
		i := int(math.Floor(ga.counter / ga.rate))
		ga.frame = ga.anims["Run"][i%len(ga.anims["Run"])]
	case jumping:
		speed := phys.vel.Y
		i := int((-speed/phys.jumpSpeed + 1) / 2 * float64(len(ga.anims["Jump"])))
		if i < 0 {
			i = 0
		}
		if i >= len(ga.anims["Jump"]) {
			i = len(ga.anims["Jump"]) - 1
		}
		ga.frame = ga.anims["Jump"][i]
	}

	// set the facing direction of the gopher
	if phys.vel.X != 0 {
		if phys.vel.X > 0 {
			ga.dir = +1
		} else {
			ga.dir = -1
		}
	}
}

func (ga *gopherAnim) draw(t pixel.Target, phys *gopherPhys) {
	if ga.sprite == nil {
		ga.sprite = pixel.NewSprite(nil, pixel.Rect{})
	}
	// draw the correct frame with the correct position and direction
	ga.sprite.Set(ga.sheet, ga.frame)
	ga.sprite.Draw(t, pixel.IM.
		ScaledXY(pixel.ZV, pixel.V(
			phys.rect.W()/ga.sprite.Frame().W(),
			phys.rect.H()/ga.sprite.Frame().H(),
		)).
		ScaledXY(pixel.ZV, pixel.V(-ga.dir, 1)).
		Moved(phys.rect.Center()),
	)
}

type hat struct {
	pos pixel.Vec
	counter float64
	color, altColor   pixel.RGBA
}

func (h *hat) update(dt float64, target pixel.Vec) {
	h.pos.X = target.X
	h.pos.Y = target.Y
	h.pos.Y += 6
}

func (h *hat) draw(imd *imdraw.IMDraw) {
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
}

type goal struct {
	pos    pixel.Vec
	radius float64
	step   float64

	counter float64
	cols    [5]pixel.RGBA
}

func (g *goal) update(dt float64, target pixel.Vec) {
	g.counter += dt
	for g.counter > g.step {
		g.counter -= g.step
		for i := len(g.cols) - 2; i >= 0; i-- {
			g.cols[i+1] = g.cols[i]
		}
		g.cols[0] = randomNiceColor()
	}

	if g.pos.X < target.X {
		g.pos.X += 0.08
	}

	if g.pos.X > target.X {
		g.pos.X -= 0.08
	}

	if g.pos.Y < target.Y {
		g.pos.Y += 0.08
	}

	if g.pos.Y > target.Y {
		g.pos.Y -= 0.08
	}
}

func (g *goal) draw(imd *imdraw.IMDraw) {
	for i := len(g.cols) - 1; i >= 0; i-- {
		imd.Color = g.cols[i]
		imd.Push(g.pos)
		imd.Circle(float64(i+1)*g.radius/float64(len(g.cols)), 0)
	}
}

type colorlight struct {
	color  pixel.RGBA
	point  pixel.Vec
	angle  float64
	radius float64
	dust   float64

	spread float64

	imd *imdraw.IMDraw
}

func randomNiceColor() pixel.RGBA {
again:
	r := rand.Float64()
	g := rand.Float64()
	b := rand.Float64()
	len := math.Sqrt(r*r + g*g + b*b)
	if len == 0 {
		goto again
	}
	return pixel.RGB(r/len, g/len, b/len)
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

func run() {
	rand.Seed(time.Now().UnixNano())

	sheet, anims, err := loadAnimationSheet("sheet.png", "sheet.csv", 12)
	if err != nil {
		panic(err)
	}

	cfg := pixelgl.WindowConfig{
		Title:  "Platformer",
		Bounds: pixel.R(0, 0, 1024, 768),
		VSync:  true,
	}
	win, err := pixelgl.NewWindow(cfg)
	if err != nil {
		panic(err)
	}

	lights := make([]colorlight, 1)
	for i := range lights {
		lights[i] = colorlight{
			color:  pixel.RGB(float64(255)/float64(255), 0, float64(250)/float64(255)),
			point:  pixel.Vec{X: win.Bounds().Center().X, Y: win.Bounds().Center().Y},
			angle:  math.Pi / 4,
			radius: 400,
			dust:   0.3,
			spread: math.Pi / math.E/2,
		}
	}

	pandaPic, err := loadPicture("panda.png")
	if err != nil {
		panic(err)
	}

	panda := pixel.NewSprite(pandaPic, pandaPic.Bounds())

	phys := &gopherPhys{
		gravity:   -512,
		runSpeed:  64,
		jumpSpeed: 192,
		rect:      pixel.R(-6, -7, 6, 7),
	}

	fire := &fire{}

	anim := &gopherAnim{
		sheet: sheet,
		anims: anims,
		rate:  1.0 / 10,
		dir:   +1,
	}

	// hardcoded level
	platforms := []platform{
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
	for i := range platforms {
		platforms[i].color = randomNiceColor()
	}

	gol := &goal{
		pos:    pixel.V(-75, 40),
		radius: 18,
		step:   1.0 / 7,
	}

	hat := &hat{color: pixel.RGB(float64(255)/float64(255), 0, float64(250)/float64(255)), altColor: pixel.RGB(float64(32)/float64(255), float64(22)/float64(255), float64(249)/float64(156))}

	canvas := pixelgl.NewCanvas(pixel.R(-160/2, -120/2, 160/2, 120/2))
	oneLight := pixelgl.NewCanvas(win.Bounds())
	allLight := pixelgl.NewCanvas(win.Bounds())

	imd := imdraw.New(sheet)
	imd.Precision = 32

	camPos := pixel.ZV

	last := time.Now()
	for !win.Closed() {
		dt := time.Since(last).Seconds()
		last = time.Now()

		// lerp the camera position towards the gopher
		camPos = pixel.Lerp(camPos, phys.rect.Center(), 1-math.Pow(1.0/128, dt))
		cam := pixel.IM.Moved(camPos.Scaled(-1))
		canvas.SetMatrix(cam)

		// slow motion with tab
		if win.Pressed(pixelgl.KeyTab) {
			dt /= 8
		}

		// restart the level on pressing enter
		if win.JustPressed(pixelgl.KeyEnter) {
			phys.rect = phys.rect.Moved(phys.rect.Center().Scaled(-1))
			phys.vel = pixel.ZV
		}

		if win.JustPressed(pixelgl.MouseButtonLeft) {
			fire.now(128, win.Bounds().Center().Add(phys.rect.Center()), win.MousePosition())
		}

		// control the gopher with keys
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

		// update the physics and animation
		phys.update(dt, ctrl, platforms)
		gol.update(dt, phys.rect.Center())
		hat.update(dt, phys.rect.Center())
		anim.update(dt, phys)

		// draw the scene to the canvas using IMDraw
		canvas.Clear(colornames.Black)
		imd.Clear()

		// accumulate all the lights
		for i := range lights {
			oneLight.Clear(pixel.Alpha(0))
			lights[i].apply(oneLight, oneLight.Bounds().Center(), panda)
			oneLight.Draw(allLight, pixel.IM.Moved(allLight.Bounds().Center()))
		}

		panda.Draw(win, pixel.IM.Moved(win.Bounds().Min))

		fire.draw(imd)
		for _, p := range platforms {
			p.draw(imd)
		}
		gol.draw(imd)
		anim.draw(imd, phys)
		hat.draw(imd)
		imd.Draw(canvas)

		// stretch the canvas to the window
		win.Clear(colornames.White)
		win.SetMatrix(pixel.IM.Scaled(pixel.ZV,
			math.Min(
				win.Bounds().W()/canvas.Bounds().W(),
				win.Bounds().H()/canvas.Bounds().H(),
			),
		).Moved(win.Bounds().Center()))
		canvas.Draw(win, pixel.IM.Moved(canvas.Bounds().Center()))

		win.Update()
	}
}

func main() {
	pixelgl.Run(run)
}

func (cl *colorlight) apply(dst pixel.ComposeTarget, center pixel.Vec, src *pixel.Sprite) {
	// create the light arc if not created already
	if cl.imd == nil {
		imd := imdraw.New(nil)
		imd.Color = pixel.Alpha(1)
		imd.Push(pixel.ZV)
		imd.Color = pixel.Alpha(0)
		for angle := -cl.spread / 2; angle <= cl.spread/2; angle += cl.spread / 64 {
			imd.Push(pixel.V(1, 0).Rotated(angle))
		}
		imd.Polygon(0)
		cl.imd = imd
	}

	// draw the light arc
	dst.SetMatrix(pixel.IM.Scaled(pixel.ZV, cl.radius).Rotated(pixel.ZV, cl.angle).Moved(cl.point))
	dst.SetColorMask(pixel.Alpha(1))
	dst.SetComposeMethod(pixel.ComposePlus)
	cl.imd.Draw(dst)

	// draw the noise inside the light
	dst.SetMatrix(pixel.IM)
	dst.SetComposeMethod(pixel.ComposeIn)
	//noise.Draw(dst, pixel.IM.Moved(center))

	// draw an image inside the noisy light
	dst.SetColorMask(cl.color)
	dst.SetComposeMethod(pixel.ComposeIn)
	src.Draw(dst, pixel.IM.Moved(center))

	// draw the light reflected from the dust
	dst.SetMatrix(pixel.IM.Scaled(pixel.ZV, cl.radius).Rotated(pixel.ZV, cl.angle).Moved(cl.point))
	dst.SetColorMask(cl.color.Mul(pixel.Alpha(cl.dust)))
	dst.SetComposeMethod(pixel.ComposePlus)
	cl.imd.Draw(dst)
}
