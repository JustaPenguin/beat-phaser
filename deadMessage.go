package main

import (
	"fmt"
	"github.com/faiface/pixel"
	"github.com/faiface/pixel/text"
	"golang.org/x/image/font/basicfont"
	"math/rand"
	"time"
)

var Messages = []string{
	"Keep Dancing",
	"Never Let the Boogie Die",
	"Take Our Energy - Use It to Dance",
}

var Colors = []pixel.RGBA{
	pixel.RGB(235.0/255.0, 35.0/255.0, 208.0/255.0),
	pixel.RGB(42.0/255.0, 251.0/255.0, 249.0/255.0),
	pixel.RGB(88.0/255.0, 16.0/255.0, 140.0/255.0),
	pixel.RGB(241.0/255.0, 169.0/255.0, 19.0/255.0),
	pixel.RGB(154.0/255.0, 192.0/255.0, 254.0/255.0),
}

type deadMessage struct {
	pos   pixel.Vec
	atlas *text.Atlas

	color pixel.RGBA

	text string

	tick     <-chan time.Time
	fastTick <-chan time.Time
}

func (d *deadMessage) init() {
	d.atlas = text.NewAtlas(
		basicfont.Face7x13,
		text.ASCII,
	)

	d.text = Messages[0]
	d.tick = time.Tick(time.Second * 2)
	d.fastTick = time.Tick(time.Millisecond * 150)
}

func (d *deadMessage) pickRandomMessage() {
	r := rand.Intn(len(Messages))

	d.text = Messages[r]
}

func (d *deadMessage) pickRandomColor() {
	r := rand.Intn(len(Colors))

	d.color = Colors[r]
}

func (d *deadMessage) update(dt float64, pos pixel.Vec) {

	select {
	case <-d.tick:
		d.pickRandomMessage()
		return
	case <-d.fastTick:
		d.pickRandomColor()
	default:

	}

	d.pos = pos
}

func (d *deadMessage) draw(t pixel.Target) {
	tx := text.New(d.pos, d.atlas)
	tx.Color = d.color

	_, err := fmt.Fprintf(tx, d.text)

	if err != nil {
		panic(err)
	}

	tx.Draw(t, pixel.IM)
}
